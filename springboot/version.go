package springboot

import (
	"fmt"
	"github.com/onsi/gomega/types"
	"golang.org/x/mod/semver"
	"strings"
)

var (
	legacyJdkVersions versions = []string{"1.5", "1.6", "1.7", "1.8"}
)

type versions []string

type versionMatcher struct {
	expect string
}

func (vs versions) match(actual string) bool {
	for _, v := range vs {
		if semver.Compare(
			semver.MajorMinor(normalize(v)),
			semver.MajorMinor(normalize(actual)),
		) == 0 {
			return true
		}
	}
	return false
}

func (v versionMatcher) Match(actual interface{}) (success bool, err error) {
	if len(strings.TrimSpace(actual.(string))) == 0 {
		return actual.(string) == v.expect, nil
	}
	return semver.Compare(
		semver.MajorMinor(normalize(v.expect)),
		semver.MajorMinor(normalize(actual.(string))),
	) == 0 || legacyJdkVersions.match(semver.MajorMinor(normalize("1."+actual.(string)))), nil
}

func normalize(version string) string {
	if len(version) == 0 {
		return version
	}
	if version[0] != 'v' {
		return "v" + SanitizeVersion(version)
	}
	return SanitizeVersion(version)
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

func LessThan(versionA, versionB string) bool {
	return semver.Compare(normalize(versionA), normalize(versionB)) < 0
}

func GreatThan(versionA, versionB string) bool {
	return semver.Compare(normalize(versionA), normalize(versionB)) > 0
}

func SanitizeVersion(version string) string {
	return strings.ReplaceAll(CleanOutput(version), "_", "-")
}

func IsValidJdkVersion(version string) bool {
	return len(semver.MajorMinor(normalize(version))) > 0
}
