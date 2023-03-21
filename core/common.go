package core

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	gookitconfig "github.com/gookit/config/v2"
	gookityaml "github.com/gookit/config/v2/yamlv3"
	"github.com/hashicorp/go-version"
	"github.com/onsi/gomega/types"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	v1alpha1 "microsoft.com/azure-spring-discovery/api/v1alpha1"
)

var (
	jdk5, _ = version.NewVersion("1.5")
	jdk6, _ = version.NewVersion("1.6")
	jdk7, _ = version.NewVersion("1.7")
	jdk8, _ = version.NewVersion("1.8")
)

const (
	KiB = float64(1024)
	MiB = float64(1024) * KiB
)

type AppType string

type AppTypes []AppType

type patterns struct {
	LoggingPatterns            []*regexp.Regexp
	AppPatterns                []*regexp.Regexp
	ConsoleOutputRegexPatterns []*regexp.Regexp
	ConsoleOutputYamlPatterns  []*yamlpath.Path
}

const (
	SpringBootFatJar   AppType = "SpringBootExecutableFatJar"
	SpringBootThinJar  AppType = "SpringBootExecutableThinJar"
	SpringBootExploded AppType = "SpringBootExploded"
	ExecutableJar      AppType = "ExecutableJar"
	Unknown            AppType = "Unknown"
)

var Patterns = NewPatterns()

func (at AppTypes) Contains(other AppType) bool {
	for _, t := range at {
		if t == other {
			return true
		}
	}
	return false
}

const (
	ManifestFile                = "MANIFEST.MF"
	JarFileExt                  = ".jar"
	AppNameField                = "Implementation-Title"
	VersionField                = "Implementation-Version"
	MainClassField              = "Main-Class"
	JdkVersionField             = "Build-Jdk-Spec"
	JdkVersionFieldFor1x        = "Build-Jdk"
	SpringBootVersionField      = "Spring-Boot-Version"
	JarLauncherClassName        = "org.springframework.boot.loader.JarLauncher"
	PropertiesLauncherClassName = "org.springframework.boot.loader.PropertiesLauncher"
	JvmOptionXmx                = "-Xmx"
	JvmOptionMaxRamPercentage   = "-XX:MaxRAMPercentage"
)

type versionMatcher struct {
	expect string
}

func (v versionMatcher) Match(actual interface{}) (success bool, err error) {
	if len(strings.TrimSpace(actual.(string))) == 0 {
		return actual.(string) == v.expect, nil
	}
	ev, err := version.NewVersion(sanitizeVersion(cleanOutput(v.expect, LinuxNewLineCharacter)))
	if err != nil {
		return false, err
	}
	av, err := version.NewVersion(actual.(string))
	if err != nil {
		return false, err
	}
	if av.Segments()[0] == jdk8.Segments()[1] ||
		av.Segments()[0] == jdk7.Segments()[1] ||
		av.Segments()[0] == jdk6.Segments()[1] ||
		av.Segments()[0] == jdk5.Segments()[1] {
		return compareVersion(ev, av) ||
			compareVersion(ev, jdk8) ||
			compareVersion(ev, jdk7) ||
			compareVersion(ev, jdk6) ||
			compareVersion(ev, jdk5), nil
	}
	return compareVersion(ev, av), nil
}

func compareVersion(v1 *version.Version, v2 *version.Version) bool {
	return v1.Segments()[0] == v2.Segments()[0] && v1.Segments()[1] == v2.Segments()[1]
}

func (v versionMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %s, but got %s", v.expect, actual)
}

func (v versionMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %s, but got %s", v.expect, actual)
}

func MatchVersion(expected string) types.GomegaMatcher {
	return versionMatcher{expect: expected}
}

func OperationIdAnn(operationId string) map[string]string {
	return map[string]string{
		v1alpha1.AnnotationsOperationId: operationId,
	}
}

func ParseYaml(content string) (*gookitconfig.Config, error) {
	cfg := gookitconfig.New("yaml-masker")
	cfg.AddDriver(gookityaml.Driver)
	cfg.WithOptions(func(opts *gookitconfig.Options) {
		opts.DumpFormat = gookitconfig.Yaml
		opts.ReadFormat = gookitconfig.Yaml
		opts.ParseKey = false
		opts.Delimiter = '.'
		opts.ParseDefault = true

	})
	err := cfg.LoadStrings(gookitconfig.Yaml, content)
	if err != nil {
		return nil, err
	}

	return cfg, nil
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

func GetConfigFromYaml[T any](key, content string) (T, bool) {
	var zero T
	cfg, err := ParseYaml(content)
	if err != nil {
		return zero, false
	}

	if find, ok := cfg.Get(key, true).(T); ok {
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

func NewPatterns() *patterns {
	var loggingPatterns []*regexp.Regexp
	var appPatterns []*regexp.Regexp
	var ps []*regexp.Regexp
	var ys []*yamlpath.Path

	for _, str := range config.GetConfigListByKey(AppConfigPatternKey) {
		appPatterns = append(appPatterns, regexp.MustCompile(str))
	}

	for _, str := range config.GetConfigListByKey(LoggingConfigPatternKey) {
		loggingPatterns = append(loggingPatterns, regexp.MustCompile(str))
	}

	for _, pattern := range ExportedConfig.GetConfigListByKey(ConsoleOutputPatternKey) {
		p := regexp.MustCompile(pattern)
		ps = append(ps, p)
	}

	for _, path := range ExportedConfig.GetConfigListByKey(ConsoleOutputYamlPathKey) {
		p, err := yamlpath.NewPath(path)
		if err != nil {
			panic(err)
		}
		ys = append(ys, p)
	}
	return &patterns{
		LoggingPatterns:            loggingPatterns,
		AppPatterns:                appPatterns,
		ConsoleOutputRegexPatterns: ps,
		ConsoleOutputYamlPatterns:  ys,
	}
}

func WithRetry(f func() error, retries int, delay time.Duration) (err error) {
	var initial = delay
	for i := 0; i < retries; i++ {
		if err = f(); err != nil {
			if !errors.As(err, &RetryableError{}) {
				return
			} else {
				time.Sleep(initial)
				initial = initial * 2
			}
		} else {
			break
		}
	}
	return
}

func Intersect[T any](left, right []T, idFunc func(t T) string, onLeft func(t T), onMid func(l, r T)) []T {
	if len(left) == 0 {
		return right
	}
	if len(right) == 0 {
		return left
	}

	leftMap := array2map(left, idFunc)
	rightMap := array2map(right, idFunc)
	var mid []T
	for key, l := range leftMap {
		if r, exists := rightMap[key]; exists {
			onMid(r, l)
			mid = append(mid, r)
			delete(leftMap, key)
			delete(rightMap, key)
		} else {
			onLeft(l)
		}
	}

	return append(mapValues(leftMap), append(mid, mapValues(rightMap)...)...)
}

func array2map[T any](arr []T, idFunc func(t T) string) map[string]T {
	var m = make(map[string]T)
	for _, item := range arr {
		key := idFunc(item)
		m[key] = item
	}
	return m
}

func mapValues[K comparable, V any](m map[K]V) []V {
	var values []V

	for _, app := range m {
		values = append(values, app)
	}

	return values
}

func mapKeys[K comparable, V any](m map[K]V) []K {
	var result []K
	for k := range m {
		result = append(result, k)
	}
	return result
}

func filterMap(m map[string]string, filterFunc func(name, content string) bool) []string {
	var result []string
	for name, content := range m {
		if filterFunc(name, content) {
			result = append(result, name)
		}
	}
	return result
}

func New[T any]() T {
	var t T
	typ := reflect.TypeOf(t)
	if typ.Kind() == reflect.Pointer {
		return reflect.New(typ.Elem()).Interface().(T)
	}
	return reflect.New(typ).Elem().Interface().(T)
}
