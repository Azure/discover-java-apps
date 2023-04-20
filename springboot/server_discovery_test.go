package springboot

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Linux java process", func() {
	var (
		ctrl               *gomock.Controller
		m                  *MockServerConnector
		executor           ServerDiscovery
		ctx                context.Context
		credentialProvider *MockCredentialProvider
		credential         *Credential
		cfg                YamlConfig
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		m = NewMockServerConnector(ctrl)
		ctx = context.Background()
		cfg = YamlCfg
		credentialProvider = NewMockCredentialProvider(ctrl)
		executor = NewLinuxServerDiscovery(ctx, m, credentialProvider, cfg)
		credential = &Credential{
			Username:       "mockuser",
			Password:       "mockpass",
			Id:             "mockid",
			FriendlyName:   "mockname",
			CredentialType: "springboot",
		}

		m.EXPECT().FQDN().Return("mock_server").AnyTimes()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Process scan", func() {
		When("got success output", func() {
			It("should return process list", func() {
				m.EXPECT().RunCmd(LinuxProcessScanCmd).Return(strings.Join([]string{ExecutableProcess, SpringBoot2xProcess}, "\n"), nil)
				processes, _ := executor.ProcessScan()
				Expect(processes).Should(HaveLen(2))

				process := processes[0]
				Expect(process.GetJvmOptions()).Should(And(
					ContainElement("-XX:InitialRAMPercentage=60.0"),
					ContainElement("-DtestOption=abc=def")),
				)
			})
		})

		When("got empty output", func() {
			It("should return empty process list", func() {
				m.EXPECT().RunCmd(LinuxProcessScanCmd).Return("", nil)
				Expect(executor.ProcessScan()).Should(BeEmpty())
			})
		})

		When("got invalid output", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(LinuxProcessScanCmd).Return(" abcdefg ", nil)
				Expect(executor.ProcessScan()).Error().Should(HaveOccurred())
			})
		})

		When("got error", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(LinuxProcessScanCmd).Return("", fmt.Errorf("test error message"))
				Expect(executor.ProcessScan()).Error().Should(HaveOccurred())
			})
		})
	})

	Context("Get os total memory", func() {
		When("got success output", func() {
			It("should return memory size in bytes", func() {
				m.EXPECT().RunCmd(LinuxGetTotalMemoryCmd).Return(TotalMemory, nil)
				Expect(executor.GetTotalMemory()).Should(Equal(int64(987654321 * KiB)))
			})
		})
		When("got empty output", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(LinuxGetTotalMemoryCmd).Return("  \n", nil)
				Expect(executor.GetTotalMemory()).Error().Should(HaveOccurred())
			})
		})
		When("got invalid output", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(LinuxGetTotalMemoryCmd).Return("  abcdefg\n", nil)
				Expect(executor.GetTotalMemory()).Error().Should(HaveOccurred())
			})
		})
		When("got error", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(LinuxGetTotalMemoryCmd).Return("", fmt.Errorf("test error message"))
				Expect(executor.GetTotalMemory()).Error().Should(HaveOccurred())
			})
		})
	})

	Context("Parse jarfile for spring boot", func() {
		var (
			b        []byte
			err      error
			checksum string
			fileInfo os.FileInfo
		)

		BeforeEach(func() {
			checksum = "test_checksum"
		})

		When("spring boot 1.x jar file read", func() {
			jar := filepath.Join("..", "mock", SpringBoot1xJarFile)
			var process JavaProcess
			BeforeEach(func() {
				b, err = os.ReadFile(jar)
				if err != nil {
					panic(err)
				}
				fileInfo, _ = os.Stat(jar)

				process = &javaProcess{
					options: []string{
						"-Dspring.application.name=test",
					},
				}
			})
			It("should be parsed as expected", func() {
				m.EXPECT().Read(gomock.Any()).Return(bytes.NewReader(b), fileInfo, nil)
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxSha256Cmd, jar)).Return(checksum, nil)
				actual, err := executor.ReadJarFile(jar, DefaultJarFileWalkers...)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(actual.GetAppType()).Should(Equal(SpringBootFatJar))
				Expect(actual.GetAppName(process)).Should(Equal("test"))
				Expect(actual.GetAppPort(process)).Should(Equal(8080))
				Expect(actual.GetChecksum()).Should(Equal(checksum))
				Expect(actual.GetSpringBootVersion()).Should(Equal("1.5.14.RELEASE"))
				Expect(actual.GetApplicationConfigurations()).Should(HaveKey("application.yaml"))
				Expect(actual.GetLoggingFiles()).Should(And(HaveLen(2), HaveKey("log4j.properties")))
				Expect(actual.GetDependencies()).Should(ContainElement("spring-boot-1.5.14.RELEASE.jar"))
				Expect(actual.GetCertificates()).Should(ContainElement("private_key.pem"))
				Expect(actual.GetStaticFiles()).Should(ContainElement("static/test.html"))
				Expect(actual.GetBuildJdkVersion()).Should(MatchVersion("1.7"))
			})
		})

		When("spring boot 2.x jar file read", func() {
			jar := filepath.Join("..", "mock", SpringBoot2xJarFile)
			var process JavaProcess
			BeforeEach(func() {
				b, err = os.ReadFile(jar)
				if err != nil {
					panic(err)
				}
				fileInfo, _ = os.Stat(jar)
				process = &javaProcess{
					options: []string{
						"-Dspring.application.name=test",
						"-Dserver.port=8085",
					},
				}
			})
			It("should be parsed as expected", func() {
				m.EXPECT().Read(gomock.Any()).Return(bytes.NewReader(b), fileInfo, nil)
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxSha256Cmd, jar)).Return("", nil)
				actual, err := executor.ReadJarFile(jar, DefaultJarFileWalkers...)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(actual.GetChecksum()).Should(Not(BeEmpty()))
				Expect(actual.GetAppType()).Should(Equal(SpringBootFatJar))
				Expect(actual.GetAppName(process)).Should(Equal("test"))
				Expect(actual.GetAppPort(process)).Should(Equal(8085))
				Expect(actual.GetSpringBootVersion()).Should(Equal("2.4.13"))
				Expect(actual.GetApplicationConfigurations()).Should(HaveKey("application.yaml"))
				Expect(actual.GetLoggingFiles()).Should(
					And(
						HaveLen(4),
						HaveKey("log4j2.xml"),
						HaveKey("log4j2.properties"),
						HaveKey("log4j2.yml"),
						HaveKey("log4j2.json"),
					),
				)
				Expect(actual.GetDependencies()).Should(ContainElement("spring-boot-2.4.13.jar"))
				Expect(actual.GetCertificates()).Should(ContainElement("private_key.pem"))
				Expect(actual.GetStaticFiles()).Should(ContainElement("static/test.html"))
				Expect(actual.GetBuildJdkVersion()).Should(MatchVersion("8"))
			})
		})

		When("executable jar file read", func() {
			jar := filepath.Join("..", "mock", ExecutableJarFile)
			var process JavaProcess
			BeforeEach(func() {
				b, err = os.ReadFile(jar)
				if err != nil {
					panic(err)
				}
				fileInfo, _ = os.Stat(jar)

				process = &javaProcess{
					options: []string{
						"-Dspring.application.name=executable_app",
						"--server.port=8075",
					},
				}
			})
			It("should be parsed as expected", func() {
				m.EXPECT().Read(gomock.Any()).Return(bytes.NewReader(b), fileInfo, nil)
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxSha256Cmd, jar)).Return(checksum, nil)
				actual, err := executor.ReadJarFile(jar, DefaultJarFileWalkers...)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(actual.GetAppType()).Should(Equal(ExecutableJar))
				Expect(actual.GetAppName(process)).Should(Equal("executable_app"))
				Expect(actual.GetAppPort(process)).Should(Equal(8075))
				Expect(actual.GetChecksum()).Should(Equal(checksum))
				Expect(actual.GetSpringBootVersion()).Should(BeEmpty())
				Expect(actual.GetApplicationConfigurations()).Should(BeEmpty())
				Expect(actual.GetLoggingFiles()).Should(BeEmpty())
				Expect(actual.GetDependencies()).Should(BeEmpty())
				Expect(actual.GetCertificates()).Should(BeEmpty())
				Expect(actual.GetStaticFiles()).Should(BeEmpty())
				Expect(actual.GetBuildJdkVersion()).Should(MatchVersion("1.7"))
			})
		})
	})

	Context("Get OS name", func() {
		It("should return as expected", func() {
			m.EXPECT().RunCmd(GetOsName()).Return("expected_os_name", nil)

			Expect(executor.GetOsName()).Should(Equal("expected_os_name"))
		})
	})

	Context("Get OS version", func() {
		It("should return as expected", func() {
			m.EXPECT().RunCmd(GetOsName()).Return("expected_os_version", nil)

			Expect(executor.GetOsName()).Should(Equal("expected_os_version"))
		})
	})

	Context("Prepare server discovery", func() {
		var (
			credentials []*Credential
		)
		BeforeEach(func() {
			executor = NewLinuxServerDiscovery(ctx, m, credentialProvider, cfg)

			for i := 0; i < 10; i++ {
				username := fmt.Sprintf("username_%d", i)
				credentials = append(credentials, &Credential{
					Username:       username,
					Password:       fmt.Sprintf("password_%d", i),
					Id:             fmt.Sprintf("id_%d", i),
					FriendlyName:   fmt.Sprintf("name_%d", i),
					CredentialType: "springboot",
				})
			}
		})

		When("connect succeeded", func() {
			It("should succeeded", func() {
				credentialProvider.EXPECT().GetCredentials().Return([]*Credential{credential}, nil).AnyTimes()
				m.EXPECT().Connect(gomock.Eq(credential.Username), gomock.Eq(credential.Password)).Return(nil)
				Expect(executor.Prepare()).Should(Equal(credential))
			})
		})

		When("credentials are empty", func() {
			It("connect should failed with credential error", func() {
				credentialProvider.EXPECT().GetCredentials().Return(nil, nil)
				m.EXPECT().Connect(gomock.Any(), gomock.Any()).MaxTimes(0)
				Expect(executor.Prepare()).Error().Should(BeAssignableToTypeOf(CredentialError{}))
			})
		})

		When("get credentials has error", func() {
			It("connect should failed with credential error", func() {
				credentialError := &CredentialError{error: fmt.Errorf("credential error")}
				credentialProvider.EXPECT().GetCredentials().Return(nil, credentialError)
				m.EXPECT().Connect(gomock.Any(), gomock.Any()).MaxTimes(0)
				Expect(executor.Prepare()).Error().Should(MatchError(credentialError))
			})
		})

		When("multiple credentials get", func() {
			It("should succeeded for one of them", func() {
				for _, cred := range credentials {
					m.EXPECT().Connect(gomock.Eq(cred.Username), gomock.Any()).MaxTimes(1)
				}
				credentialProvider.EXPECT().GetCredentials().Return(credentials, nil)
				Expect(executor.Prepare()).Should(BeElementOf(credentials))
			})
		})

		When("multiple credentials get, when slow connect", func() {
			It("should succeeded finally", func() {
				for _, cred := range credentials {
					call := m.EXPECT().Connect(gomock.Eq(cred.Username), gomock.Any()).MaxTimes(1)
					slowCall(call)
				}
				credentialProvider.EXPECT().GetCredentials().Return(credentials, nil)
				Expect(executor.Prepare()).Should(BeElementOf(credentials))
			})
		})

		When("multiple credentials get, when error occurred for all", func() {
			It("should failed", func() {
				for _, cred := range credentials {
					call := m.EXPECT().Connect(gomock.Eq(cred.Username), gomock.Any()).MaxTimes(1)
					errorCall(call)
				}
				credentialProvider.EXPECT().GetCredentials().Return(credentials, nil)
				Expect(executor.Prepare()).Error().Should(Not(BeNil()))
			})
		})

		When("multiple credentials get, when error occurred in partial", func() {
			It("should succeeded", func() {
				for i, cred := range credentials {
					call := m.EXPECT().Connect(gomock.Eq(cred.Username), gomock.Any()).MaxTimes(1)
					if i%2 == 0 {
						errorCall(call)
					}
				}
				credentialProvider.EXPECT().GetCredentials().Return(credentials, nil)
				Expect(executor.Prepare()).Error().Should(BeNil())
			})
		})

		When("multiple credentials get, when auth error occurred", func() {
			It("should succeeded", func() {
				for i, cred := range credentials {
					call := m.EXPECT().Connect(gomock.Eq(cred.Username), gomock.Any()).MaxTimes(1)
					if i%2 == 0 {
						unauthenticated(call)
					}
				}
				credentialProvider.EXPECT().GetCredentials().Return(credentials, nil)
				Expect(executor.Prepare()).Error().Should(BeNil())
			})
		})
	})

	Context("Finish server discovery", func() {
		When("connect succeeded", func() {
			It("should succeeded", func() {
				m.EXPECT().Close().MinTimes(1)
				Expect(executor.Finish()).Should(Succeed())
			})
		})
	})

})

func slowCall(call *gomock.Call) *gomock.Call {
	return call.DoAndReturn(func(username, password string) error {
		time.Sleep(5 * time.Second)
		return nil
	})
}

func errorCall(call *gomock.Call) *gomock.Call {
	return call.Return(fmt.Errorf("connection refused"))
}

func unauthenticated(call *gomock.Call) *gomock.Call {
	return call.Return(fmt.Errorf("ssh: unable to authenticate"))
}
