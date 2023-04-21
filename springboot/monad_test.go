package springboot

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test monadic", func() {

	Context("test monad", func() {

		var (
			monadic *Monadic[*SpringBootApp]
			process JavaProcess
			jar     JarFile
			appName string
		)

		BeforeEach(func() {
			appName = "test-app"
			monadic = NewMonadic[*SpringBootApp](process, jar)
		})

		When("happy path", func() {
			It("should succeeded without err", func() {
				monadic.Apply(func(process JavaProcess, jarFile JarFile) *Monad {
					return Of(appName, nil).Field("AppName")
				})

				app, err := monadic.Get()

				Expect(app).Should(Not(BeNil()))
				Expect(err).Should(BeNil())
				Expect(app.AppName).Should(Equal(appName))
			})
		})

		When("error path", func() {
			It("should failed with err", func() {
				monadic.Apply(func(process JavaProcess, jarFile JarFile) *Monad {
					return Of("", fmt.Errorf("cannot get app name")).Field("AppName")
				})

				app, err := monadic.Get()

				Expect(app).Should(BeNil())
				Expect(err).Should(Not(BeNil()))
			})
		})

		When("error path 2", func() {
			It("should failed with err", func() {
				monadic.Apply(func(process JavaProcess, jarFile JarFile) *Monad {
					return Of(SpringBootFatJar, nil).Field("AppType")
				})
				monadic.Apply(func(process JavaProcess, jarFile JarFile) *Monad {
					return Of("", fmt.Errorf("cannot get app name")).Field("AppName")
				})

				app, err := monadic.Get()

				Expect(app).Should(BeNil())
				Expect(err).Should(Not(BeNil()))
			})
		})

		When("filed not exists", func() {
			It("should return empty", func() {
				monadic.Apply(func(process JavaProcess, jarFile JarFile) *Monad {
					return Of(appName, nil).Field("NotExistsField")
				})

				app, err := monadic.Get()

				Expect(app).Should(Not(BeNil()))
				Expect(err).Should(BeNil())
			})
		})

	})
})
