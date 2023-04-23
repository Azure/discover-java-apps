package springboot

import (
	"github.com/golang/mock/gomock"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JarFile test", func() {
	var (
		ctrl           *gomock.Controller
		j              *jarFile
		process        *javaProcess
		defaultAppName string
		defaultAppPort int
	)
	BeforeEach(func() {
		defaultAppName = "testapp"
		defaultAppPort = 8083
		ctrl = gomock.NewController(GinkgoT())
		process = &javaProcess{
			options: []string{},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Get config from manifests", func() {
		When("manifest is not empty", func() {
			BeforeEach(func() {
				process.options = []string{
					"--server.port=" + strconv.Itoa(defaultAppPort),
					"-Dspring.application.name=" + defaultAppName,
				}
				j = &jarFile{manifests: parseManifests(Manifest)}
			})

			It("should return spring boot version", func() {
				Expect(j.GetSpringBootVersion()).Should(Equal("2.4.13"))
			})

			It("should return app name", func() {
				Expect(j.GetAppName(process)).Should(Equal(defaultAppName))
			})

			It("should return artifact name", func() {
				Expect(j.GetArtifactName()).Should(Equal("hellospring"))
			})

			It("should return app port", func() {
				Expect(j.GetAppPort(process)).Should(Equal(defaultAppPort))
			})

			It("should return build jdk version", func() {
				Expect(j.GetBuildJdkVersion()).Should(Equal("11"))
			})
		})

		When("manifest is empty", func() {
			BeforeEach(func() {
				j = &jarFile{manifests: parseManifests(""), remoteLocation: "hellospringfromfilename.jar"}
				process = &javaProcess{
					options: []string{
						"-Dserver.port=8084",
					},
				}
			})

			It("should return empty spring boot version", func() {
				Expect(j.GetSpringBootVersion()).Should(BeEmpty())
			})

			It("should return jar file name as fallback", func() {
				Expect(j.GetAppName(process)).Should(Equal("hellospringfromfilename"))
			})

			It("should return artifact name", func() {
				Expect(j.GetArtifactName()).Should(Equal("hellospringfromfilename"))
			})

			It("should return app port", func() {
				Expect(j.GetAppPort(process)).Should(Equal(8084))
			})

			It("should return empty build jdk version", func() {
				Expect(j.GetBuildJdkVersion()).Should(BeEmpty())
			})
		})
	})

	Context("Get config from pom", func() {
		When("manifest is not empty", func() {
			BeforeEach(func() {
				process.options = []string{
					"-Dserver.port=" + strconv.Itoa(defaultAppPort),
				}
				mvnProject, err := readPom(Pom)
				if err != nil {
					panic(err)
				}
				j = &jarFile{mvnProject: mvnProject, applicationConfigurations: map[string]string{
					"application.properties": "spring.application.name=hellospringfromprops",
				}}
			})

			It("should return spring boot version", func() {
				Expect(j.GetSpringBootVersion()).Should(Equal("2.4.13"))
			})

			It("should return app name from application.properties", func() {
				Expect(j.GetAppName(process)).Should(Equal("hellospringfromprops"))
			})

			It("should return artifact name", func() {
				Expect(j.GetArtifactName()).Should(Equal("hellospring"))
			})

			It("should return app port", func() {
				Expect(j.GetAppPort(process)).Should(Equal(defaultAppPort))
			})

			It("should return build jdk version", func() {
				Expect(j.GetBuildJdkVersion()).Should(Equal("8"))
			})
		})

		When("manifest is empty", func() {
			BeforeEach(func() {
				j = &jarFile{manifests: parseManifests(""), remoteLocation: "hellospringtest.jar", applicationConfigurations: map[string]string{
					"application.yaml": "spring:\n  application:\n    name: hellospringfromyaml\nserver:\n  port: 8083",
				}}
			})

			It("should return empty spring boot version", func() {
				Expect(j.GetSpringBootVersion()).Should(BeEmpty())
			})

			It("should return app name from application.yaml", func() {
				Expect(j.GetAppName(process)).Should(Equal("hellospringfromyaml"))
			})

			It("should return artifact name", func() {
				Expect(j.GetArtifactName()).Should(Equal("hellospringtest"))
			})

			It("should return app port from yaml", func() {
				Expect(j.GetAppPort(process)).Should(Equal(8083))
			})

			It("should return empty build jdk version", func() {
				Expect(j.GetBuildJdkVersion()).Should(BeEmpty())
			})
		})
	})
})
