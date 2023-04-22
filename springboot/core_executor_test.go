package springboot

import (
	"bytes"
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"os"
	"path/filepath"
	"strings"
)

var osName = "ubuntu"
var osVersion = "1804"
var fqdn = "test-fqdn"
var testUser = "test-user"

var _ = Describe("Test springboot discovery executor", func() {
	var (
		executor               DiscoveryExecutor
		credentialProvider     *MockCredentialProvider
		serverConnectorFactory *MockServerConnectorFactory
		serverConnector        *MockServerConnector
		credentials            []*Credential
		ctrl                   *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		credentialProvider = NewMockCredentialProvider(ctrl)
		serverConnectorFactory = NewMockServerConnectorFactory(ctrl)
		serverConnector = NewMockServerConnector(ctrl)
		executor = NewSpringBootDiscoveryExecutor(credentialProvider, serverConnectorFactory, YamlCfg)
		credentials = []*Credential{
			{
				Username: testUser,
				Password: "password",
			},
		}
		format.MaxLength = 0
	})

	When("server is accessible", func() {
		It("apps should be discovered as expected", func() {
			credentialProvider.EXPECT().GetCredentials().Return(credentials, nil).AnyTimes()
			connection := ServerConnectionInfo{Server: fqdn, Port: 1022}
			serverConnector.EXPECT().FQDN().Return(fqdn).AnyTimes()
			serverConnector.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			serverConnectorFactory.EXPECT().Create(gomock.Any(), fqdn, gomock.Any()).Return(serverConnector).AnyTimes()

			var matchers []types.GomegaMatcher
			var processes []string
			for _, testcase := range []struct {
				appName           string
				jdkVersion        string
				springBootVersion string
				jarFileLocation   string
				process           string
			}{
				{appName: ExecutableAppName, jdkVersion: Jdk7Version, springBootVersion: "", jarFileLocation: ExecutableJarFileLocation, process: ExecutableProcess},
				{appName: SpringBoot1xAppName, jdkVersion: Jdk7Version, springBootVersion: SpringBoot1xVersion, jarFileLocation: SpringBoot1xJarFileLocation, process: SpringBoot1xProcess},
				{appName: SpringBoot2xAppName, jdkVersion: Jdk7Version, springBootVersion: SpringBoot2xVersion, jarFileLocation: SpringBoot2xJarFileLocation, process: SpringBoot2xProcess},
			} {
				if testcase.appName != ExecutableAppName {
					matchers = append(matchers, MatchFatJar(testcase.appName, testcase.jdkVersion, testcase.springBootVersion, testcase.jarFileLocation))
				} else {
					matchers = append(matchers, MatchExecutableJar(testcase.appName, testcase.jarFileLocation))
				}
				processes = append(processes, testcase.process)
			}

			setupServerConnectorMock(serverConnector, strings.Join(processes, "\n"))
			apps, err := executor.Discover(context.Background(), connection)
			Expect(apps).Should(ContainElements(matchers))
			Expect(err).Should(BeNil())
		})
	})

	When("server is not accessible", func() {
		It("discovery should be failed", func() {
			serverConnectorFactory.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(serverConnector).AnyTimes()
			credentialProvider.EXPECT().GetCredentials().Return(credentials, nil).AnyTimes()
			serverConnector.EXPECT().FQDN().Return(fqdn).AnyTimes()
			serverConnector.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(fmt.Errorf("connection error")).AnyTimes()
			serverConnector.EXPECT().Close().MinTimes(1)

			connection := ServerConnectionInfo{
				Server: fqdn,
				Port:   1022,
			}
			apps, err := executor.Discover(context.Background(), connection)

			Expect(apps).Should(BeNil())
			Expect(err).Should(Not(BeNil()))
		})
	})

	When("primary fqdn is not accessible but alternative IP address is accessible", func() {
		It("prepare should succeeded", func() {
			var primaryFqdn = "primary"
			var altServerAccessible = "server-accessible"
			var altServerNotAccessible = "server-not-accessible"

			for _, testcase := range []struct {
				fqdn       string
				accessible bool
			}{
				{fqdn: primaryFqdn, accessible: false},
				{fqdn: altServerAccessible, accessible: true},
				{fqdn: altServerNotAccessible, accessible: false},
			} {
				connector := NewMockServerConnector(ctrl)
				connector.EXPECT().FQDN().Return(testcase.fqdn).AnyTimes()
				if testcase.accessible {
					connector.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				} else {
					connector.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(fmt.Errorf("connection error")).AnyTimes()
				}
				setupServerConnectorMock(connector, strings.Join([]string{SpringBoot2xProcess, SpringBoot1xProcess, ExecutableProcess}, "\n"))
				serverConnectorFactory.EXPECT().Create(gomock.Any(), testcase.fqdn, gomock.Any()).Return(connector).AnyTimes()
			}

			credentialProvider.EXPECT().GetCredentials().Return(credentials, nil).AnyTimes()

			connection := ServerConnectionInfo{Server: primaryFqdn, Port: 1022}
			accessible := ServerConnectionInfo{Server: altServerAccessible, Port: 1022}
			nonaccessible := ServerConnectionInfo{Server: altServerNotAccessible, Port: 1022}
			apps, err := executor.Discover(context.Background(), connection, nonaccessible, accessible)

			Expect(apps).Should(Not(BeEmpty()))
			Expect(err).Should(BeNil())
		})
	})

	When("get credentials error", func() {
		It("discovery should be failed", func() {
			serverConnectorFactory.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(serverConnector).AnyTimes()
			credentialProvider.EXPECT().GetCredentials().Return(nil, fmt.Errorf("get credential error")).AnyTimes()
			serverConnector.EXPECT().FQDN().Return(fqdn).AnyTimes()
			serverConnector.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			serverConnector.EXPECT().Close().MinTimes(1)

			connection := ServerConnectionInfo{
				Server: fqdn,
				Port:   1022,
			}
			apps, err := executor.Discover(context.Background(), connection)

			Expect(apps).Should(BeNil())
			Expect(err).Should(Not(BeNil()))
		})
	})

})

func MatchFatJar(app string, jdkVersion string, springBootVersion string, jarLocation string) types.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"AppName": Equal(app),
		"AppType": Equal(SpringBootFatJar),
		"Artifact": PointTo(MatchFields(IgnoreExtras, Fields{
			"Group":   Not(BeEmpty()),
			"Name":    Not(BeEmpty()),
			"Version": Not(BeEmpty()),
		})),
		"Runtime": PointTo(MatchFields(IgnoreExtras, Fields{
			"Server":            Equal(fqdn),
			"Uid":               BeNumerically(">=", 0),
			"Pid":               BeNumerically(">=", 0),
			"AppPort":           Equal(8080),
			"JavaCmd":           Not(BeEmpty()),
			"RuntimeJdkVersion": MatchVersion("11"),
			"Environments": And(
				ContainElement(Equal("test_option=test")),
				ContainElement(Equal("DB_PASSWORD=testpassword1234")),
			),
			"JvmOptions": And(
				ContainElement(Or(Equal("-XX:InitialRAMPercentage"), Equal("-Xmx128m"))),
				ContainElement(Equal("-Dcom.sun.management.jmxremote.password=testpassword1234")),
				ContainElement(Equal("-javaagent:/path/to/applicationinsights.jar")),
			),
			"JvmMemory":    BeNumerically(">", 0),
			"BindingPorts": Not(BeEmpty()),
			"OsName":       Equal(osName),
			"OsVersion":    Equal(osVersion),
		})),
		"SpringBootVersion": Equal(springBootVersion),
		"BuildJdkVersion":   MatchVersion(jdkVersion),
		"Checksum":          Not(BeEmpty()),
		"Dependencies":      Or(ContainElement("spring-boot-2.4.13.jar"), ContainElement("spring-boot-1.5.14.RELEASE.jar")),
		"ApplicationConfigurations": And(
			HaveKeyWithValue(Equal("application.yaml"), Or(ContainSubstring("credential"), ContainSubstring("datasource"))),
			HaveKeyWithValue(Equal("application.properties"), Or(ContainSubstring("server.port"), ContainSubstring("spring.application.name"))),
		),
		"LoggingConfigurations":  Not(BeEmpty()),
		"Certificates":           ContainElement("private_key.pem"),
		"JarFileLocation":        Equal(jarLocation),
		"StaticContentLocations": ContainElement("/other_static"),
		"LastUpdatedTime":        Not(BeNil()),
		"LastModifiedTime":       Not(BeNil()),
	}))
}

func MatchExecutableJar(app string, jarLocation string) types.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"AppName": Equal(app),
		"AppType": Equal(ExecutableJar),
		"Artifact": PointTo(MatchFields(IgnoreExtras, Fields{
			"Group":   Not(BeEmpty()),
			"Name":    Not(BeEmpty()),
			"Version": Not(BeEmpty()),
		})),
		"Runtime": PointTo(MatchFields(IgnoreExtras, Fields{
			"Server":            Equal(fqdn),
			"Uid":               Equal(1000),
			"Pid":               Equal(ExecutableProcessId),
			"AppPort":           Equal(8080),
			"JavaCmd":           Equal("java"),
			"RuntimeJdkVersion": MatchVersion("11"),
			"Environments": And(
				ContainElement(Equal("test_option=test")),
				ContainElement(Equal("DB_PASSWORD=testpassword1234")),
			),
			"JvmOptions": And(
				ContainElement(Or(Equal("-XX:InitialRAMPercentage=60.0"), Equal("-Xmx128m"))),
				ContainElement(Equal("-Dcom.sun.management.jmxremote.password=testpassword1234")),
				ContainElement(Equal("-javaagent:/path/to/applicationinsights.jar")),
			),
			"JvmMemory":    BeNumerically(">", 0),
			"BindingPorts": Not(BeEmpty()),
			"OsName":       Equal(osName),
			"OsVersion":    Equal(osVersion),
		})),
		"SpringBootVersion":         BeEmpty(),
		"BuildJdkVersion":           MatchVersion("1.7"),
		"Dependencies":              BeEmpty(),
		"ApplicationConfigurations": BeEmpty(),
		"LoggingConfigurations":     BeEmpty(),
		"Certificates":              BeEmpty(),
		"JarFileLocation":           Equal(jarLocation),
		"StaticContentLocations":    BeEmpty(),
		"LastUpdatedTime":           Not(BeNil()),
		"LastModifiedTime":          Not(BeNil()),
	}))
}

func setupServerConnectorMock(s *MockServerConnector, processes string) {
	s.EXPECT().Close().MinTimes(1)
	s.EXPECT().RunCmd(gomock.Eq(GetLocateJarCmd(SpringBoot2xProcessId, SpringBoot2xJarFile))).Return(SpringBoot2xJarFileLocation, nil).AnyTimes()
	s.EXPECT().RunCmd(gomock.Eq(GetEnvCmd(SpringBoot2xProcessId))).Return(TestEnv, nil).AnyTimes()
	s.EXPECT().RunCmd(gomock.Eq(GetPortsCmd(SpringBoot2xProcessId))).Return(Ports, nil).AnyTimes()

	s.EXPECT().RunCmd(gomock.Eq(GetLocateJarCmd(SpringBoot1xProcessId, SpringBoot1xJarFile))).Return(SpringBoot1xJarFileLocation, nil).AnyTimes()
	s.EXPECT().RunCmd(gomock.Eq(GetEnvCmd(SpringBoot1xProcessId))).Return(TestEnv, nil).AnyTimes()
	s.EXPECT().RunCmd(gomock.Eq(GetPortsCmd(SpringBoot1xProcessId))).Return(Ports, nil).AnyTimes()

	s.EXPECT().RunCmd(gomock.Eq(GetLocateJarCmd(ExecutableProcessId, ExecutableJarFile))).Return(ExecutableJarFileLocation, nil).AnyTimes()
	s.EXPECT().RunCmd(gomock.Eq(GetEnvCmd(ExecutableProcessId))).Return(TestEnv, nil).AnyTimes()
	s.EXPECT().RunCmd(gomock.Eq(GetPortsCmd(ExecutableProcessId))).Return(Ports, nil).AnyTimes()

	s.EXPECT().RunCmd(gomock.Eq(GetProcessScanCmd())).Return(processes, nil).AnyTimes()

	s.EXPECT().RunCmd(CmdMatcher(LinuxGetTotalMemoryCmd)).Return(TotalMemory, nil).AnyTimes()
	s.EXPECT().RunCmd(CmdMatcher(LinuxGetJdkVersionCmd)).Return(RuntimeJdkVersion, nil).AnyTimes()
	s.EXPECT().RunCmd(CmdMatcher(LinuxGetDefaultMaxHeapCmd)).Return(DefaultMaxHeapSize, nil).AnyTimes()
	s.EXPECT().RunCmd(CmdMatcher(LinuxSha256Cmd)).Return("", nil).AnyTimes()
	s.EXPECT().RunCmd(CmdMatcher(GetOsName())).Return(osName, nil).AnyTimes()
	s.EXPECT().RunCmd(CmdMatcher(GetOsVersion())).Return(osVersion, nil).AnyTimes()
	//s.EXPECT().FQDN().Return(Host).AnyTimes()

	b, err := os.ReadFile(filepath.Join("..", "mock", SpringBoot2xJarFile))
	if err != nil {
		panic(err)
	}
	info, _ := os.Stat(filepath.Join("..", "mock", SpringBoot2xJarFile))
	s.EXPECT().Read(gomock.Eq(SpringBoot2xJarFileLocation)).Return(bytes.NewReader(b), info, nil).AnyTimes()

	b, err = os.ReadFile(filepath.Join("..", "mock", SpringBoot1xJarFile))
	if err != nil {
		panic(err)
	}
	info, _ = os.Stat(filepath.Join("..", "mock", SpringBoot1xJarFile))
	s.EXPECT().Read(gomock.Eq(SpringBoot1xJarFileLocation)).Return(bytes.NewReader(b), info, nil).AnyTimes()

	b, err = os.ReadFile(filepath.Join("..", "mock", ExecutableJarFile))
	if err != nil {
		panic(err)
	}
	info, _ = os.Stat(filepath.Join("..", "mock", ExecutableJarFile))
	s.EXPECT().Read(gomock.Eq(ExecutableJarFileLocation)).Return(bytes.NewReader(b), info, nil).AnyTimes()
}
