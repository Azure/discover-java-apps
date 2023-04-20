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

			result, _ := ToSlice[string](s.Take(2))
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

			result := s.Reduce(0, func(a, b any) any {
				return a.(int) + b.(int)
			})

			Expect(result).Should(Equal(37))
		})
	})

	Context("test parallel stream", func() {

		It("should be as expected", func() {
			s := FromSlice(context.Background(), []int{7, 4, 8, 3, 5, 1, 9})
			s = s.Parallel(3)

			result := s.Filter(
				func(t any) bool {
					return t.(int) > 2
				},
			).Sorted(
				func(a, b any) any {
					return a.(int) - b.(int)
				},
			).Take(
				5,
			).Map(
				func(context context.Context, i int) (string, error) {
					return fmt.Sprintf("%v", i), nil
				},
			).Map(
				func(context context.Context, i string) string {
					return fmt.Sprintf("%v", i)
				},
			).Reduce("test",
				func(a, b any) any {
					return a.(string) + "," + b.(string)
				},
			)

			Expect(result).Should(Equal("test,3,4,5,7,8"))
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
				Expect(result).Should(Equal([]int{9, 8, 7, 5, 4, 3, 1}))

				Expect(s.Join("=")).Should(Equal("9=8=7=5=4=3=1"))
			})
		})

		When("error occurred with retry enabled", func() {
			It("should failed with retried", func() {
				s := FromSlice(context.Background(), []int{1, 3, 6, 7})
				s = s.Parallel(3)

				m := make(map[int]int)
				mux := sync.Mutex{}
				s = s.Retry(
					PolicyOf(3, time.Millisecond*500),
				).Map(
					func(context context.Context, i int) (int, error) {
						mux.Lock()
						m[i]++
						mux.Unlock()
						if i%3 == 0 {
							return 0, fmt.Errorf("error")
						}
						return i, nil
					},
				)

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

				By(fmt.Sprintf("running test case in gid %v", gid))
				s := FromSlice(context.Background(), [][]string{
					{"1", "2", "3"},
					{"3", "4", "5"},
					{"5", "6", "7"},
				})
				s = s.Parallel(3)
				s.FlatMap(func(t []string) Stream {
					return FromSlice[string](context.Background(), t)
				}).Map(
					func(s string) (int64, error) {
						Expect(getGID()).Should(Not(Equal(gid)))
						return strconv.ParseInt(s, 10, 32)
					},
				).Distinct().Filter(func(t any) bool {
					Expect(getGID()).Should(Not(Equal(gid)))
					return t.(int64) > 3
				}).Take(5).Peek(func(t int64) {
					Expect(getGID()).Should(Not(Equal(gid)))
				}).Sorted(func(a, b int64) int {
					Expect(getGID()).Should(Equal(gid)) // for a parallel stream, the sorting is in main thread
					return int(a - b)
				}).ForEach(func(i int64) {
					Expect(getGID()).Should(Equal(gid)) // after sorting in main thread, forEach is also in main thread
				})
			})
		})
	})

})
