package springboot

import (
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"regexp"
)

var Patterns = newPatterns()

type patterns struct {
	LoggingPatterns            []*regexp.Regexp
	AppPatterns                []*regexp.Regexp
	ConsoleOutputRegexPatterns []*regexp.Regexp
	MavenPomVersionPattern     *regexp.Regexp
	ConsoleOutputYamlPatterns  []*yamlpath.Path
}

func newPatterns() *patterns {
	var loggingPatterns []*regexp.Regexp
	var appPatterns []*regexp.Regexp
	var ps []*regexp.Regexp
	var ys []*yamlpath.Path

	for _, str := range YamlCfg.Pattern.App {
		appPatterns = append(appPatterns, regexp.MustCompile(str))
	}

	for _, str := range YamlCfg.Pattern.Logging.FilePatterns {
		loggingPatterns = append(loggingPatterns, regexp.MustCompile(str))
	}

	for _, pattern := range YamlCfg.Pattern.Logging.ConsoleOutput.Patterns {
		p := regexp.MustCompile(pattern)
		ps = append(ps, p)
	}

	for _, path := range YamlCfg.Pattern.Logging.ConsoleOutput.Yamlpath {
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
		MavenPomVersionPattern:     regexp.MustCompile("(-?[0-9\\.]+.*)?\\.jar"),
	}
}
