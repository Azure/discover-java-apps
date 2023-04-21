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
	unsupported        unaryFuncType = iota
	zeroInSingleOut                  // f() any
	zeroInErrorOut                   // f() (any, error)
	singleInZeroOut                  // f(any)
	singleInSingleOut                // f(any) any
	singleInErrorOut                 // f(any) (any, error)
	contextInZeroOut                 // f(context, any)
	contextInSingleOut               // f(context, any) any
	contextInErrorOut                // f(context, any) (any, error)
)

var stopped = stop{}

func FromSlice[T any](ctx context.Context, slice []T) Stream {
	return stream(func(consumer consumer) error {
		var err error
		for _, i := range slice {
			err = consumer(ctx, i)
			if err != nil {
				return err
			}
		}
		return nil
	})
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

type Stream interface {
	consume(consumer consumer) error
	asyncConsume(consumer consumer, parallelism int) error
	ForEach(f interface{}) error
	Peek(f interface{}) Stream
	Map(f interface{}) Stream
	FlatMap(f interface{}) Stream
	Distinct() Stream
	Filter(f Predicate) Stream
	Join(sep string) (string, error)
	Retry(policy RetryPolicy) Stream
	Take(n int) Stream
	First() (any, error)
	Sorted(less interface{}) Stream
	Reduce(seed any, c Combinator[any]) (any, error)
	Parallel(parallelism int) Stream
	GroupBy(keyFunc func(t any) string) (map[string][]any, error)
}

type stop struct {
}

func (s stop) Error() string {
	return "stopped"
}

type Comparator[T comparable] func(a, b T) int

type Combinator[T any] func(a, b T) T

type Predicate func(t any) bool

type consumer func(ctx context.Context, t any) error

type stream func(consumer consumer) error

type RetryPolicy struct {
	max      int
	interval time.Duration
}

func (s stream) consume(c consumer) error {
	return s(c)
}

func (s stream) asyncConsume(c consumer, parallelism int) error {
	errs, _ := errgroup.WithContext(context.Background())
	errs.SetLimit(parallelism)

	_ = s.ForEach(func(t any) {
		errs.Go(func() error {
			return c(context.Background(), t)
		})
	})
	return errs.Wait()
}

func (s stream) ForEach(f interface{}) error {
	fc := toCall(f)
	return s(func(ctx context.Context, t any) error {
		_, err := fc.call(ctx, t)
		return err
	})
}

func (s stream) Peek(f interface{}) Stream {
	fc := toCall(f)
	return s.Map(func(ctx context.Context, t any) (any, error) {
		_, err := fc.call(ctx, t)
		return t, err
	})
}

func (s stream) Map(f interface{}) Stream {
	fc := toCall(f)
	return stream(func(consumer consumer) error {
		return s(func(ctx context.Context, t any) error {
			val, err := fc.call(ctx, t)
			if err != nil {
				return err
			}
			return consumer(ctx, val.Interface())
		})
	})
}

func (s stream) FlatMap(f interface{}) Stream {
	fc := toCall(f)
	return stream(func(consumer consumer) error {
		return s(func(ctx context.Context, t any) error {
			val, err := fc.call(ctx, t)
			if err != nil {
				return err
			}
			return val.Interface().(Stream).consume(consumer)
		})
	})
}

func (s stream) Distinct() Stream {
	return stream(func(consumer consumer) error {
		var m = make(map[any]bool)
		var mux sync.Mutex
		return s.consume(func(ctx context.Context, t any) error {
			mux.Lock()
			_, exists := m[t]
			if !exists {
				m[t] = true
			}
			mux.Unlock()
			if !exists {
				return consumer(ctx, t)
			}
			return nil
		})
	})
}

func (s stream) Filter(f Predicate) Stream {
	return stream(func(consumer consumer) error {
		return s(func(ctx context.Context, t any) error {
			if f(t) {
				return consumer(ctx, t)
			}
			return nil
		})
	})
}

func (s stream) Join(sep string) (string, error) {
	var slice []string
	var mux sync.Mutex
	err := s.ForEach(func(t any) {
		mux.Lock()
		defer mux.Unlock()
		slice = append(slice, fmt.Sprintf("%v", t))
	})
	if err != nil && err != stopped {
		return "", err
	}

	return strings.Join(slice, sep), nil
}

func (s stream) Retry(policy RetryPolicy) Stream {
	return stream(func(consumer consumer) error {
		return s(func(ctx context.Context, t any) error {
			var err error
			for i := 0; i <= policy.max; i++ {
				err = consumer(ctx, t)
				if err == nil || err == stopped {
					break
				}
				time.Sleep(policy.interval)
			}
			return err
		})
	})
}

func (s stream) Take(n int) Stream {
	return stream(func(consumer consumer) error {
		i := n
		var mux sync.Mutex
		return s(func(ctx context.Context, t any) error {
			shouldConsume := false
			mux.Lock()
			i--
			shouldConsume = i >= 0
			mux.Unlock()
			if shouldConsume {
				return consumer(ctx, t)
			} else {
				return stopped
			}
		})
	})
}

func (s stream) First() (any, error) {
	var result any
	var mux sync.Mutex
	err := s(func(ctx context.Context, t any) error {
		mux.Lock()
		defer mux.Unlock()
		result = t
		return stopped
	})
	return result, ignoreStopped(err)
}

func (s stream) Sorted(less interface{}) Stream {
	return stream(func(consumer consumer) error {
		var slice []any
		var ctx context.Context
		var mux sync.Mutex
		_ = s(func(context context.Context, t any) error {
			mux.Lock()
			defer mux.Unlock()
			slice = append(slice, t)
			ctx = context
			return nil
		})

		sortSlice(slice, less)

		for _, t := range slice {
			err := consumer(ctx, t)
			if err != nil && err != stopped {
				return err
			}
		}
		return nil
	})
}

func (s stream) Reduce(initial any, c Combinator[any]) (any, error) {
	var result = initial
	var mux sync.Mutex
	err := s.ForEach(func(t any) {
		mux.Lock()
		defer mux.Unlock()
		result = c(result, t)
	})
	return result, ignoreStopped(err)
}

func (s stream) Parallel(parallelism int) Stream {
	return stream(func(consumer consumer) error {
		return s.asyncConsume(consumer, parallelism)
	})
}

func (s stream) GroupBy(keyFunc func(t any) string) (map[string][]any, error) {
	var m = make(map[string][]any)
	var mux sync.Mutex
	err := s(func(ctx context.Context, t any) error {
		mux.Lock()
		defer mux.Unlock()
		key := keyFunc(t)
		if _, exists := m[key]; exists {
			m[key] = append(m[key], t)
		} else {
			m[key] = []any{t}
		}
		return nil
	})

	if err != nil && err != stopped {
		return nil, err
	}
	return m, nil
}

func isUnaryFunc(refType reflect.Type) unaryFuncType {
	if refType.Kind() != reflect.Func {
		return unsupported
	}

	switch refType.NumIn() {
	case 0:
		switch refType.NumOut() {
		case 0:
			return unsupported
		case 1:
			return zeroInSingleOut
		case 2:
			if isError(refType.Out(1)) {
				return zeroInErrorOut
			}
			return unsupported
		}
	case 1:
		switch refType.NumOut() {
		case 0:
			return singleInZeroOut
		case 1:
			return singleInSingleOut
		case 2:
			if isError(refType.Out(1)) {
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
			if isError(refType.Out(1)) {
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
		return result, fmt.Errorf("unsupported function " + refVal.String())
	case zeroInSingleOut:
		out := refVal.Call(toReflectParams())
		if hasError(out[0]) {
			return result, out[0].Interface().(error)
		}
		return out[0], nil
	case zeroInErrorOut:
		out := refVal.Call(toReflectParams())
		if hasError(out[1]) {
			return result, out[1].Interface().(error)
		}
		return out[0], nil
	case singleInZeroOut:
		refVal.Call(toReflectParams(data))
	case singleInSingleOut:
		out := refVal.Call(toReflectParams(data))
		if hasError(out[0]) {
			return result, out[0].Interface().(error)
		}
		return out[0], nil
	case singleInErrorOut:
		out := refVal.Call(toReflectParams(data))
		if out[1].Interface() == nil {
			return out[0], nil
		}
		return result, out[1].Interface().(error)
	case contextInZeroOut:
		refVal.Call(toReflectParams(ctx, data))
	case contextInSingleOut:
		out := refVal.Call(toReflectParams(ctx, data))
		if hasError(out[0]) {
			return result, out[0].Interface().(error)
		}
		return out[0], nil
	case contextInErrorOut:
		out := refVal.Call(toReflectParams(ctx, data))
		if out[1].Interface() == nil {
			return out[0], nil
		}
		return result, out[1].Interface().(error)
	}

	return result, nil
}

func ToSlice[T any](s Stream) ([]T, error) {
	var slice []T
	var err error
	var mux sync.Mutex
	err = s.consume(func(ctx context.Context, t any) error {
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

func sortSlice(slice []any, less interface{}) {
	sort.Slice(slice, func(i, j int) bool {
		ret := reflect.ValueOf(less).Call(
			[]reflect.Value{
				reflect.ValueOf(slice[i]),
				reflect.ValueOf(slice[j]),
			},
		)
		return ret[0].Interface().(int) < 0
	})
}

type funcCall struct {
	refTyp unaryFuncType
	refVal reflect.Value
}

func toCall(f interface{}) funcCall {
	return funcCall{
		refTyp: isUnaryFunc(reflect.TypeOf(f)),
		refVal: reflect.ValueOf(f),
	}
}

func (c funcCall) call(ctx context.Context, t any) (reflect.Value, error) {
	return callUnaryFunc(ctx, c.refVal, t, c.refTyp)
}

func ignoreStopped(err error) error {
	if err == stopped {
		return nil
	}
	return err
}

func isError(typ reflect.Type) bool {
	return typ == reflect.TypeOf((*error)(nil)).Elem()
}

func hasError(value reflect.Value) bool {
	return isError(value.Type()) && !value.IsNil()
}

func toReflectParams(params ...any) []reflect.Value {
	var vals []reflect.Value
	for _, t := range params {
		vals = append(vals, reflect.ValueOf(t))
	}

	return vals
}
