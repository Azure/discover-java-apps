package springboot

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("test version matcher", func() {

	When("normal version matched", func() {
		It("should as expected", func() {
			m := versionMatcher{expect: "1.8"}

			Expect(m.Match("1.8.0")).Should(BeTrue())
		})
	})

	When("version with prerelease matched", func() {
		It("should as expected", func() {
			m := versionMatcher{expect: "1.8"}

			Expect(m.Match("1.8.231_ea")).Should(BeTrue())
		})
	})

	When("java 8 version matched", func() {
		It("should as expected", func() {
			m := versionMatcher{expect: "1.8"}

			Expect(m.Match("8")).Should(BeTrue())
		})
	})

	When("invalid version matched", func() {
		It("should not matched", func() {
			m := versionMatcher{expect: "1.8"}

			Expect(m.Match("a.b.c")).Should(BeFalse())
		})
	})

	When("greater than", func() {
		It("should success", func() {
			Expect(GreatThan("1.8", "1.7")).Should(BeTrue())
			Expect(GreatThan("1.8.232_ea", "1.7")).Should(BeTrue())
			Expect(GreatThan("1.8", "1.7.24")).Should(BeTrue())
			Expect(GreatThan("1.7", "1.7.24")).Should(BeFalse())
			Expect(GreatThan("1.7.24", "1.7")).Should(BeTrue())
			Expect(GreatThan("a.b.c", "1.7")).Should(BeFalse())
		})
	})

	When("less than", func() {
		It("should success", func() {
			Expect(LessThan("1.7", "1.8")).Should(BeTrue())
			Expect(LessThan("1.7", "1.8.232_ea")).Should(BeTrue())
			Expect(LessThan("1.7.24", "1.8")).Should(BeTrue())
			Expect(LessThan("1.7.24", "1.7")).Should(BeFalse())
			Expect(LessThan("1.7", "1.7.24")).Should(BeTrue())
			Expect(LessThan("a.b.c", "1.7")).Should(BeTrue())
		})
	})

	When("is valid jdk version", func() {
		It("should return as expected", func() {
			Expect(IsValidJdkVersion("a.b.c.d")).Should(BeFalse())
			Expect(IsValidJdkVersion("1.b.c.d")).Should(BeFalse())
			Expect(IsValidJdkVersion("1.7.c.d")).Should(BeFalse())
			Expect(IsValidJdkVersion("1.7.6.d")).Should(BeFalse())
			Expect(IsValidJdkVersion("1.7.6-d")).Should(BeTrue())
			Expect(IsValidJdkVersion("1.7.6_d")).Should(BeTrue())
		})
	})
})
