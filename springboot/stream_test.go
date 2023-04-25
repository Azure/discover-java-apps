package springboot

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strconv"
	"sync"
	"time"
)

var _ = Describe("Stream API test", func() {

	Context("test map", func() {

		It("int map to string", func() {
			s := FromSlice(context.Background(), []int{1, 2, 3})

			result, err := ToSlice[string](s.Map(func(i int) string {
				return fmt.Sprintf("%v", i)
			}))

			Expect(result).Should(ConsistOf("1", "2", "3"))
			Expect(err).Should(BeNil())
		})

		It("string map to int", func() {
			s := FromSlice(context.Background(), []string{"1", "2", "3"})

			result, err := ToSlice[int64](s.Map(func(i string) (int64, error) {
				return strconv.ParseInt(i, 10, 32)
			}))

			Expect(result).Should(ConsistOf(int64(1), int64(2), int64(3)))
			Expect(err).Should(BeNil())
		})

		It("string map to int with conversion error", func() {
			s := FromSlice(context.Background(), []string{"1", "a", "3"})

			result, err := ToSlice[int64](s.Map(func(i string) (int64, error) {
				return strconv.ParseInt(i, 10, 32)
			}))

			Expect(result).Should(BeNil())
			Expect(err).Should(Not(BeNil()))
		})
	})

	Context("test filter", func() {

		It("filter int array", func() {
			s := FromSlice(context.Background(), []int{1, 2, 3})

			result, err := ToSlice[int](s.Filter(func(i any) bool {
				return i.(int) >= 2
			}))

			Expect(result).Should(ConsistOf(3, 2))
			Expect(err).Should(BeNil())
		})

		It("filter int array out of range", func() {
			s := FromSlice(context.Background(), []int{1, 2, 3})

			result, err := ToSlice[int](s.Filter(func(i any) bool {
				return i.(int) >= 4
			}))

			Expect(result).Should(BeEmpty())
			Expect(err).Should(BeNil())
		})
	})

	Context("test for each", func() {
		It("should walk through each one", func() {
			s := FromSlice(context.Background(), []int{1, 2, 3})

			m := make(map[int]bool)
			err := s.ForEach(func(i any) {
				m[i.(int)] = true
			})

			Expect(m).Should(
				And(
					HaveKeyWithValue(1, true),
					HaveKeyWithValue(2, true),
					HaveKeyWithValue(3, true),
				),
			)
			Expect(err).Should(BeNil())
		})
	})

	Context("test peek", func() {
		It("should walk through each one", func() {
			s := FromSlice(context.Background(), []int{1, 2, 3})

			m := make(map[int]bool)
			_, err := ToSlice[int](s.Peek(func(i any) {
				m[i.(int)] = true
			}))

			Expect(m).Should(
				And(
					HaveKeyWithValue(1, true),
					HaveKeyWithValue(2, true),
					HaveKeyWithValue(3, true),
				),
			)
			Expect(err).Should(BeNil())
		})
	})

	Context("test flat map", func() {
		It("should flat map to slice", func() {
			s := FromSlice(context.Background(), [][]int{
				{1, 2, 3},
				{3, 4, 5},
				{5, 6, 7},
			})

			result, err := ToSlice[int](s.FlatMap(
				func(ctx context.Context, i []int) Stream {
					return FromSlice[int](ctx, i)
				}),
			)

			Expect(result).Should(ConsistOf(1, 2, 3, 3, 4, 5, 5, 6, 7))
			Expect(err).Should(BeNil())
		})
	})

	Context("test distinct", func() {
		It("should distinct", func() {
			s := FromSlice[int](context.Background(), []int{1, 2, 3, 3, 4, 5, 5, 6, 7})

			result, err := ToSlice[int](s.Distinct())

			Expect(result).Should(ConsistOf(1, 2, 3, 4, 5, 6, 7))
			Expect(err).Should(BeNil())
		})
	})

	Context("test string join", func() {
		It("should joined", func() {
			s := FromSlice(context.Background(), []string{"1", "a", "3"})

			Expect(s.Join(",")).Should(Equal("1,a,3"))

		})
	})

	Context("test take", func() {
		It("should return first n elements", func() {
			s := FromSlice(context.Background(), []string{"1", "a", "3"})

			result, err := ToSlice[string](s.Take(2))
			Expect(err).Should(BeNil())
			Expect(result).Should(ConsistOf("1", "a"))
		})
	})

	Context("test first", func() {
		It("should return first element", func() {
			s := FromSlice(context.Background(), []string{"1", "a", "3"})

			Expect(s.First()).Should(Equal("1"))
		})
	})

	Context("test sorted", func() {
		It("should slice in order", func() {
			s := FromSlice(context.Background(), []int{7, 4, 8, 3, 5, 1, 9})

			result, _ := ToSlice[int](s.Sorted(func(i, j int) int {
				return i - j
			}))

			Expect(result).Should(Equal([]int{1, 3, 4, 5, 7, 8, 9}))
		})
	})

	Context("test reduce", func() {
		It("should sum as expected", func() {
			s := FromSlice(context.Background(), []int{7, 4, 8, 3, 5, 1, 9})

			result, err := s.Reduce(0, func(a, b any) any {
				return a.(int) + b.(int)
			})

			Expect(result).Should(Equal(37))
			Expect(err).Should(BeNil())
		})
	})

	Context("test group by", func() {
		When("int stream supplied, count duplicated number", func() {
			It("should be grouped", func() {
				s := FromSlice(context.Background(), []int{1, 2, 3, 3, 4, 5, 5, 6, 7})

				keyFunc := func(t any) string { return fmt.Sprintf("%d", t) }
				result, err := s.GroupBy(keyFunc)

				Expect(result).Should(
					And(
						HaveKeyWithValue("1", ConsistOf(1)),
						HaveKeyWithValue("2", ConsistOf(2)),
						HaveKeyWithValue("3", ConsistOf(3, 3)),
						HaveKeyWithValue("4", ConsistOf(4)),
						HaveKeyWithValue("5", ConsistOf(5, 5)),
						HaveKeyWithValue("6", ConsistOf(6)),
						HaveKeyWithValue("7", ConsistOf(7)),
					),
				)
				Expect(err).Should(BeNil())
			})
		})
	})

	Context("test parallel stream", func() {

		It("should be as expected", func() {
			s := FromSlice(context.Background(), []int{7, 4, 8, 3, 5, 1, 9})
			s = s.Parallel(3)
			take := 5
			initial := "test"
			result, err := s.Filter(
				func(t any) bool {
					return t.(int) > 2
				},
			).Sorted(
				func(a, b any) any {
					return a.(int) - b.(int)
				},
			).Take(
				take,
			).Map(
				func(context context.Context, i int) (string, error) {
					return fmt.Sprintf("%v", i), nil
				},
			).Map(
				func(context context.Context, i string) string {
					return fmt.Sprintf("%v", i)
				},
			).Reduce(initial,
				func(a, b any) any {
					return a.(string) + "," + b.(string)
				},
			)

			Expect(result).Should(And(HavePrefix(initial), HaveLen(len(initial)+take*2)))
			Expect(err).Should(BeNil())
		})

		When("distinct in parallel", func() {
			It("should distinct", func() {
				s := FromSlice(context.Background(), []int{7, 1, 5, 4, 8, 3, 5, 7, 1, 9})
				s = s.Parallel(3)
				s = s.Distinct().Sorted(
					func(a, b any) any {
						return b.(int) - a.(int)
					},
				)
				result, _ := ToSlice[int](s)
				expected := []int{9, 8, 7, 5, 4, 3, 1}
				Expect(result).Should(ConsistOf(expected))

				Expect(s.Join("=")).Should(HaveLen(len(expected)*2 - 1))
			})
		})

		When("error occurred with retry enabled", func() {
			It("should failed with retried", func() {
				s := FromSlice(context.Background(), []int{1, 3, 6, 7})
				s = s.Parallel(50)

				m := make(map[int]int)
				mux := sync.Mutex{}
				s = s.Retry(PolicyOf(3, time.Millisecond*500)).
					Map(func(context context.Context, i int) (int, error) {
						mux.Lock()
						m[i]++
						mux.Unlock()
						if i%3 == 0 {
							return 0, fmt.Errorf("error")
						}
						return i, nil
					})

				_, err := ToSlice[int](s)

				Expect(m).Should(And(
					HaveKeyWithValue(1, 1),
					HaveKeyWithValue(3, 4),
					HaveKeyWithValue(6, 4),
					HaveKeyWithValue(7, 1),
				))
				Expect(err).Should(Not(BeNil()))
			})
		})

		When("error occurred", func() {
			It("should be failed", func() {
				s := FromSlice(context.Background(), []string{"1", "a", "3"})
				s = s.Parallel(3)

				_, err := ToSlice[int64](s.Map(func(i string) (int64, error) {
					return strconv.ParseInt(i, 10, 32)
				}))

				Expect(err).Should(Not(BeNil()))
			})
		})

		When("ensure it's parallel", func() {
			It("should succeeded", func() {
				gid := getGID()
				slice := [][]string{
					{"1", "2", "3", "4"},
					{"3", "4", "5", "6"},
					{"5", "6", "7", "8"},
					{"7", "8", "9", "10"},
				}
				By(fmt.Sprintf("running test case in gid %v", gid))
				s := FromSlice(context.Background(), slice)
				s = s.Parallel(50)
				Expect(getGID()).Should(Equal(gid)) // after sorting in main thread, forEach is also in main thread
				s = s.Parallel(50)
				for _, is := range []struct {
					name   string
					stream Stream
					sync   bool
				}{
					{name: "FlatMap", stream: s.FlatMap(func(t []string) Stream { return FromSlice[string](context.Background(), t) }).Distinct()},
					{name: "Retry", stream: s.Retry(PolicyOf(3, time.Second))},
					{name: "Map", stream: s.Map(func(s any) (any, error) { return s, nil })},
					{name: "FlatMap-Map", stream: s.FlatMap(func(s []string) Stream { return FromSlice(context.Background(), s) }).Map(func(s any) (any, error) { return s, nil })},
					{name: "Filter", stream: s.Filter(func(t any) bool { return true })},
					{name: "Take", stream: s.Take(5)},
					{name: "Peek", stream: s.Peek(func(t any) {})},
					{name: "Sorted", stream: s.Sorted(func(a, b any) int { return 1 }), sync: true},
				} {
					By(fmt.Sprintf("checking %s when ForEach", is.name))
					start := time.Now()
					_ = is.stream.ForEach(func(i any) {
						time.Sleep(time.Second)
						if is.sync {
							Expect(getGID()).Should(Equal(gid))
						} else {
							Expect(getGID()).ShouldNot(Equal(gid))
						}
					})
					Expect(time.Since(start)).Should(MatchDurationAround(time.Second*time.Duration(len(slice)), time.Millisecond*50))
				}
			})
		})
	})

	Context("test error handling", func() {

		var (
			s Stream
		)

		BeforeEach(func() {
			s = FromSlice[string](context.Background(), []string{"1", "2", "3"})
		})

		When("test multiple cases", func() {
			It("should succeeded", func() {
				for _, testcase := range []struct {
					name    string
					f       interface{}
					isError bool
				}{
					{name: "unsupported 1", f: func(a, b string) string { return a }, isError: true},
					{name: "unsupported 2", f: func(a string) (string, int) { return "", 0 }, isError: true},
					{name: "zero in single out 1", f: func() string { return "string" }, isError: false},
					{name: "zero in single out 2", f: func() error { return fmt.Errorf("error") }, isError: true},
					{name: "zero in error out 1", f: func() (string, error) { return "string", nil }, isError: false},
					{name: "zero in error out 2", f: func() (string, error) { return "", fmt.Errorf("error") }, isError: true},
					{name: "single in zero out", f: func(s string) {}, isError: false},
					{name: "single in single out 1", f: func(s string) string { return s }, isError: false},
					{name: "single in single out 2", f: func(s string) error { return fmt.Errorf("err") }, isError: true},
					{name: "single in error out 1", f: func(s string) (string, error) { return s, nil }, isError: false},
					{name: "single in error out 2", f: func(s string) (string, error) { return s, fmt.Errorf("err") }, isError: true},
					{name: "context in zero out", f: func(ctx context.Context, s string) {}, isError: false},
					{name: "context in single out 1", f: func(ctx context.Context, s string) string { return s }, isError: false},
					{name: "context in single out 2", f: func(ctx context.Context, s string) error { return fmt.Errorf("err") }, isError: true},
					{name: "context in error out 1", f: func(ctx context.Context, s string) (string, error) { return s, nil }, isError: false},
					{name: "context in error out 2", f: func(ctx context.Context, s string) (string, error) { return s, fmt.Errorf("err") }, isError: true},
				} {
					err := s.ForEach(testcase.f)
					if !testcase.isError {
						Expect(err).Should(BeNil())
					} else {
						Expect(err).ShouldNot(BeNil())
					}
				}
			})
		})
	})

})
