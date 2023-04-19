package springboot

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/golang/mock/gomock"
	"gopkg.in/yaml.v3"
	"reflect"
	"regexp"
	"strings"
)

func New[T any]() T {
	var t T
	typ := reflect.TypeOf(t)
	if typ.Kind() == reflect.Pointer {
		return reflect.New(typ.Elem()).Interface().(T)
	}
	return reflect.New(typ).Elem().Interface().(T)
}

func ParseYaml(content string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	err := yaml.Unmarshal([]byte(content), m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func ParseProperties(content string) map[string]string {
	s := bufio.NewScanner(bytes.NewBufferString(content))
	var m = make(map[string]string)
	for s.Scan() {
		line := s.Text()
		idx := strings.Index(line, "=")
		if idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+1:])
			m[key] = value
		} else {
			m[line] = ""
		}
	}
	return m
}

func flatten(m map[string]interface{}) map[string]interface{} {
	var result = make(map[string]interface{})
	var f func(parent string, m map[string]interface{}, result map[string]interface{})
	f = func(parent string, m map[string]interface{}, result map[string]interface{}) {
		for k, v := range m {
			var key string
			if len(parent) > 0 {
				key = strings.Join([]string{parent, k}, ".")
			} else {
				key = k
			}
			switch v.(type) {
			case map[string]interface{}:
				f(key, v.(map[string]interface{}), result)
			default:
				result[key] = v
			}
		}
	}
	f("", m, result)
	return result
}

func GetConfigFromYaml[T any](key, content string) (T, bool) {
	var zero T
	cfg, err := ParseYaml(content)
	if err != nil {
		return zero, false
	}
	cfg = flatten(cfg)
	if find, ok := cfg[key].(T); ok {
		return find, true
	}

	return zero, false
}

func GetConfigFromProperties(key, content string) (string, bool) {
	cfg := ParseProperties(content)

	if find, ok := cfg[key]; ok {
		return find, len(find) > 0
	}

	return "", false
}

func CleanOutput(raw string) string {
	var value = raw
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "\r\n")
	value = strings.TrimSuffix(value, "\n")
	value = strings.ReplaceAll(value, "\"", "")
	value = strings.ReplaceAll(value, "'", "")
	return value
}

func Contains(slices []string, find string) bool {
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

func CmdMatcher(cmd string) gomock.Matcher {
	return &cmdMatcher{cmd: cmd}
}
