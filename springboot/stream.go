package springboot

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type unaryFuncType int

const (
	unsupported unaryFuncType = iota
	singleInZeroOut
	singleInSingleOut
	singleInErrorOut
	contextInZeroOut
	contextInSingleOut
	contextInErrorOut
)

var stopped = stop{}

type Stream interface {
	consume(consumer consumer) error
	ForEach(f interface{}) error
	Peek(f interface{}) Stream
	Map(f interface{}) Stream
	FlatMap(f func(ctx context.Context, t any) Stream) Stream
	Distinct() Stream
	Filter(f Predicate) Stream
	Parallel(parallelism int) Stream
	Sequential() Stream
	Join(sep string) (string, error)
	Retry(policy RetryPolicy) Stream
	Take(n int) Stream
	First() (any, error)
	Sorted(less interface{}) Stream
	Reduce(seed any, c Combinator[any]) any
}

type stop struct {
}

func (s stop) Error() string {
	return "stopped"
}

type Producer func() (any, error)

type Comparator[T comparable] func(a, b T) int

type Combinator[T any] func(a, b T) T

type Predicate func(t any) bool

type consumer func(ctx context.Context, t any) error

type stream func(consumer consumer) error

type streamCall struct {
	stream
	parallel bool
}

type parallelStream struct {
	Stream
}

type RetryPolicy struct {
	max      int
	interval time.Duration
}

func IntComparator() Comparator[int] {
	return func(a, b int) int {
		return a - b
	}
}

func StringComparator() Comparator[string] {
	return func(a, b string) int {
		return strings.Compare(a, b)
	}
}

func IntSum() Combinator[any] {
	return func(a, b any) any {
		return a.(int) + b.(int)
	}
}

func StringJoiner(sep string) Combinator[any] {
	return func(a, b any) any {
		return a.(string) + sep + b.(string)
	}
}

func FromProducer(ctx context.Context, p Producer) Stream {
	return streamCall{
		stream: func(consumer consumer) error {
			t, err := p()
			if err != nil {
				return err
			}
			return consumer(ctx, t)
		},
	}
}

func FromSlice[T any](ctx context.Context, slice []T) Stream {
	return streamCall{
		stream: func(consumer consumer) error {
			var err error
			for _, i := range slice {
				err = consumer(ctx, i)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func (s streamCall) consume(consumer consumer) error {
	return s.stream(consumer)
}

func (s stream) ForEach(f interface{}) error {
	refTyp := isUnaryFunc(reflect.TypeOf(f))
	refVal := reflect.ValueOf(f)
	return s(func(ctx context.Context, t any) error {
		_, err := callUnaryFunc(ctx, refVal, t, refTyp)
		return err
	})
}

func (s streamCall) Peek(f interface{}) Stream {
	refTyp := isUnaryFunc(reflect.TypeOf(f))
	refVal := reflect.ValueOf(f)

	return streamCall{
		stream: func(consumer consumer) error {
			return s.consume(func(ctx context.Context, t any) error {
				_, err := callUnaryFunc(ctx, refVal, t, refTyp)
				if err != nil {
					return err
				}
				err = consumer(ctx, t)
				if err == stopped {
					return nil
				}
				return err
			})
		},
	}
}

func (s streamCall) Map(f interface{}) Stream {
	refTyp := isUnaryFunc(reflect.TypeOf(f))
	refVal := reflect.ValueOf(f)
	return streamCall{
		parallel: s.parallel,
		stream: func(consumer consumer) error {
			return s.consume(func(ctx context.Context, t any) error {
				val, err := callUnaryFunc(ctx, refVal, t, refTyp)
				if err != nil {
					return err
				}
				err = consumer(ctx, val.Interface())
				if err == stopped {
					return nil
				}
				return err
			})
		},
	}
}

func (s streamCall) FlatMap(f func(ctx context.Context, t any) Stream) Stream {
	return streamCall{
		stream: func(consumer consumer) error {
			return s.consume(func(ctx context.Context, t any) error {
				err := f(ctx, t).consume(consumer)
				if err == stopped {
					return nil
				}
				return err
			})
		},
	}
}

func (s streamCall) Distinct() Stream {
	if s.parallel {
		return streamCall{
			parallel: true,
			stream: func(consumer consumer) error {
				var m sync.Map
				return s.consume(func(ctx context.Context, t any) error {
					if _, ok := m.LoadOrStore(t, true); !ok {
						return consumer(ctx, t)
					}
					return nil
				})
			},
		}
	}
	return streamCall{
		stream: func(consumer consumer) error {
			var m = make(map[any]bool)
			return s.consume(func(ctx context.Context, t any) error {
				if _, ok := m[t]; !ok {
					m[t] = true
					err := consumer(ctx, t)
					if err == stopped {
						return nil
					}
					return err
				}
				return nil
			})
		},
	}
}

func (s streamCall) Filter(f Predicate) Stream {
	return streamCall{
		parallel: s.parallel,
		stream: func(consumer consumer) error {
			return s.consume(func(ctx context.Context, t any) error {
				if f(t) {
					err := consumer(ctx, t)
					if err == stopped {
						return nil
					}
					return err
				}
				return nil
			})
		},
	}
}

func (s streamCall) Parallel(parallelism int) Stream {
	return streamCall{
		parallel: true,
		stream: func(consumer consumer) (err error) {
			errs, _ := errgroup.WithContext(context.Background())
			errs.SetLimit(parallelism)
			defer func() {
				err = errs.Wait()
				if err == stopped {
					err = nil
				}
			}()
			_ = s.consume(func(ctx context.Context, t any) error {
				errs.Go(func() error {
					return consumer(ctx, t)
				})
				return nil
			})

			return nil
		},
	}
}

func (s streamCall) Sequential() Stream {
	return streamCall{
		stream:   s.stream,
		parallel: false,
	}
}

func (s streamCall) Join(sep string) (string, error) {
	if s.parallel {
		var slice []string
		var mux sync.Mutex
		err := s.consume(func(ctx context.Context, t any) error {
			mux.Lock()
			defer mux.Unlock()
			slice = append(slice, fmt.Sprintf("%v", t))
			return nil
		})

		if err != nil && err != stopped {
			return "", err
		}

		return strings.Join(slice, sep), nil
	}

	var slice []string
	err := s.consume(func(ctx context.Context, t any) error {
		slice = append(slice, fmt.Sprintf("%v", t))
		return nil
	})
	if err != nil && err != stopped {
		return "", err
	}

	return strings.Join(slice, sep), nil
}

func (s streamCall) Retry(policy RetryPolicy) Stream {
	return streamCall{
		stream: func(consumer consumer) error {
			return s.consume(func(ctx context.Context, t any) error {
				var err error
				for i := 0; i < policy.max; i++ {
					err = consumer(ctx, t)
					if err == nil || err == stopped {
						break
					}
					time.Sleep(policy.interval)
				}
				return err
			})
		},
	}
}

func (s streamCall) Take(n int) Stream {
	if s.parallel {
		return streamCall{
			parallel: true,
			stream: func(consumer consumer) error {
				var mux sync.Mutex
				i := n
				return s.consume(func(ctx context.Context, t any) error {
					shouldConsume := false
					mux.Lock()
					i--
					shouldConsume = i >= 0
					mux.Unlock()
					if shouldConsume {
						err := consumer(ctx, t)
						if err == stopped {
							return nil
						}
						return err
					} else {
						return stopped
					}
				})
			},
		}
	}
	return streamCall{
		stream: func(consumer consumer) error {
			i := n
			return s.consume(func(ctx context.Context, t any) error {
				i--
				if i >= 0 {
					err := consumer(ctx, t)
					if err == stopped {
						return nil
					}
					return err
				} else {
					return stopped
				}
			})
		},
	}
}

func (s streamCall) First() (any, error) {
	if s.parallel {
		var result any
		var mux sync.Mutex
		err := s.consume(func(ctx context.Context, t any) error {
			mux.Lock()
			defer mux.Unlock()
			result = t
			return stopped
		})
		if err == stopped {
			return result, nil
		}
		return result, err
	}
	var result any
	err := s.consume(func(ctx context.Context, t any) error {
		result = t
		return stopped
	})
	if err == stopped {
		return result, nil
	}
	return result, err
}

func (s streamCall) Sorted(less interface{}) Stream {
	if s.parallel {
		return streamCall{
			parallel: true,
			stream: func(consumer consumer) error {
				var slice []any
				var ctx context.Context
				var mux sync.Mutex
				_ = s.consume(func(context context.Context, t any) error {
					mux.Lock()
					defer mux.Unlock()
					ctx = context
					slice = append(slice, t)
					return nil
				})

				sort.Slice(slice, func(i, j int) bool {
					ret := reflect.ValueOf(less).Call(
						[]reflect.Value{
							reflect.ValueOf(slice[i]),
							reflect.ValueOf(slice[j]),
						},
					)
					return ret[0].Interface().(int) < 0
				})

				for _, t := range slice {
					err := consumer(ctx, t)
					if err != nil && err != stopped {
						return err
					}
				}
				return nil
			},
		}
	}
	return streamCall{
		stream: func(consumer consumer) error {
			var slice []any
			var ctx context.Context
			_ = s.consume(func(context context.Context, t any) error {
				ctx = context
				slice = append(slice, t)
				return nil
			})

			sort.Slice(slice, func(i, j int) bool {
				ret := reflect.ValueOf(less).Call(
					[]reflect.Value{
						reflect.ValueOf(slice[i]),
						reflect.ValueOf(slice[j]),
					},
				)
				return ret[0].Interface().(int) < 0
			})

			for _, t := range slice {
				err := consumer(ctx, t)
				if err != nil && err != stopped {
					return err
				}
			}
			return nil
		},
	}
}

func (s streamCall) Reduce(initial any, c Combinator[any]) any {
	if s.parallel {
		var result = initial
		var mux sync.Mutex
		_ = s.consume(func(ctx context.Context, t any) error {
			mux.Lock()
			defer mux.Unlock()
			result = c(result, t)
			return nil
		})

		return result
	}
	var result = initial
	_ = s.consume(func(ctx context.Context, t any) error {
		result = c(result, t)
		return nil
	})

	return result
}

func isUnaryFunc(refType reflect.Type) unaryFuncType {
	if refType.Kind() != reflect.Func {
		return unsupported
	}

	switch refType.NumIn() {
	case 1:
		switch refType.NumOut() {
		case 0:
			return singleInZeroOut
		case 1:
			return singleInSingleOut
		case 2:
			if refType.Out(1).Kind() == reflect.Interface {
				return singleInErrorOut
			}
			return unsupported
		}
	case 2:
		param0 := refType.In(0)
		if param0.Kind() != reflect.Interface {
			return unsupported
		}
		switch refType.NumOut() {
		case 0:
			return contextInZeroOut
		case 1:
			return contextInSingleOut
		case 2:
			if refType.Out(1).Kind() == reflect.Interface {
				return contextInErrorOut
			}
			return unsupported
		}
	}

	return unsupported
}

func callUnaryFunc(ctx context.Context, refVal reflect.Value, data interface{}, typ unaryFuncType) (reflect.Value, error) {
	var result reflect.Value
	switch typ {
	case unsupported:
		panic("unsupported function " + refVal.String() + ", only f(any)/f(any)error/f(any)(any,error)/f(ctx, any)/f(ctx, any)error/f(ctx, any)(any, error) supported")
	case singleInZeroOut:
		refVal.Call([]reflect.Value{reflect.ValueOf(data)})
	case singleInSingleOut:
		out := refVal.Call([]reflect.Value{reflect.ValueOf(data)})
		return out[0], nil
	case singleInErrorOut:
		out := refVal.Call([]reflect.Value{reflect.ValueOf(data)})
		if out[1].Interface() == nil {
			return out[0], nil
		}
		return result, out[1].Interface().(error)
	case contextInZeroOut:
		refVal.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(data)})
	case contextInSingleOut:
		out := refVal.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(data)})
		return out[0], nil
	case contextInErrorOut:
		out := refVal.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(data)})
		if out[1].Interface() == nil {
			return out[0], nil
		}
		return result, out[1].Interface().(error)
	}

	return result, nil
}

func ToSlice[T any](stream Stream) ([]T, error) {
	var slice []T
	var err error
	var mux sync.Mutex
	err = stream.consume(func(ctx context.Context, t any) error {
		mux.Lock()
		defer mux.Unlock()
		slice = append(slice, t.(T))
		return nil
	})
	if err != nil && err != stopped {
		return nil, err
	}
	return slice, err
}

func PolicyOf(max int, interval time.Duration) RetryPolicy {
	return RetryPolicy{max: max, interval: interval}
}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, true)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}