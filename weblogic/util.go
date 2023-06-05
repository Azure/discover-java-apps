package weblogic

import (
	"fmt"
	"regexp"
	"strings"
)

func CleanOutput(raw string) string {
	var value = raw
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "\r\n")
	value = strings.TrimSuffix(value, "\n")
	value = strings.ReplaceAll(value, "\"", "")
	value = strings.ReplaceAll(value, "'", "")
	return value
}

func Contains[T ~string](slices []T, find T) bool {
	for _, t := range slices {
		if t == find {
			return true
		}
	}
	return false
}

type tryFunc[In any, Out any] func(in In) (Out, bool)

type tryFuncs[In any, Out any] []tryFunc[In, Out]

func (ts tryFuncs[In, Out]) try(in In) (Out, bool) {
	var zero Out
	for _, f := range ts {
		if value, ok := f(in); ok {
			return value, true
		}
	}

	return zero, false
}

type cmdMatcher struct {
	cmd string
}

func (c cmdMatcher) Matches(x interface{}) bool {
	s := x.(string)
	pattern := regexp.QuoteMeta(c.cmd)
	pattern = strings.ReplaceAll(pattern, "%%d", "^^d") // a dirty trick, we need to replace %% to ^^ to keep it as-is
	pattern = strings.ReplaceAll(pattern, "%d", "[0-9]+")
	pattern = strings.ReplaceAll(pattern, "%\\[1\\]d", "[0-9]+")
	pattern = strings.ReplaceAll(pattern, "%f", "[0-9\\.]+")
	pattern = strings.ReplaceAll(pattern, "%s", "[0-9a-zA-Z\\-_\\./]+")
	pattern = strings.ReplaceAll(pattern, "^^d", "%d")
	r := regexp.MustCompile(pattern)
	find := r.FindString(s)
	return strings.Compare(s, find) == 0
}

func (c cmdMatcher) String() string {
	return fmt.Sprintf("has cmd: %s", c.cmd)
}
