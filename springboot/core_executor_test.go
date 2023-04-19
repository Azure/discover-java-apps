package springboot

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Azure/discover-java-apps/mock"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"os"
	"path/filepath"
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
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())
		credentialProvider = NewMockCredentialProvider(ctrl)
		serverConnectorFactory = NewMockServerConnectorFactory(ctrl)
		serverConnector = NewMockServerConnector(ctrl)
		executor = NewSpringBootDiscoveryExecutor(
			credentialProvider,
			serverConnectorFactory,
			YamlCfg,
		)
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
			serverConnectorFactory.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(serverConnector).AnyTimes()
			credentialProvider.EXPECT().GetCredentials().Return(credentials, nil).AnyTimes()
			serverConnector.EXPECT().FQDN().Return(fqdn).AnyTimes()
			serverConnector.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

			setupServerConnectorMock(serverConnector)

			connection := ServerConnectionInfo{
				Server: fqdn,
				Port:   1022,
			}
			apps, err := executor.Discover(context.Background(), connection)

			Expect(apps).Should(
				ContainElement(
					MatchApp(mock.SpringBoot1xAppName, mock.Jdk7Version, mock.SpringBoot1xVersion, mock.SpringBoot1xJarFileLocation, 1),
				),
			)
			Expect(err).Should(BeNil())
		})
	})

	When("server is not accessible", func() {
		It("discovery should be failed", func() {
			serverConnectorFactory.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(serverConnector).AnyTimes()
			credentialProvider.EXPECT().GetCredentials().Return(credentials, nil).AnyTimes()
			serverConnector.EXPECT().FQDN().Return(fqdn).AnyTimes()
			serverConnector.EXPECT().Connect(gomock.Any(), gomock.Any()).Return(fmt.Errorf("connection error")).AnyTimes()

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
			var altServerAccessible = "server-accessible"
			var altServerNotAccessible = "server-not-accessible"

			ctrl := gomock.NewController(GinkgoT())

			accessibleConnector := NewMockServerConnector(ctrl)
			notAccessibleConnector := NewMockServerConnector(ctrl)
			serverConnectorFactory.EXPECT().Create(gomock.Any(), fqdn, gomock.Any()).Return(serverConnector).AnyTimes()
			serverConnectorFactory.EXPECT().Create(gomock.Any(), altServerAccessible, gomock.Any()).Return(accessibleConnector).AnyTimes()
			serverConnectorFactory.EXPECT().Create(gomock.Any(), altServerNotAccessible, gomock.Any()).Return(notAccessibleConnector).AnyTimes()
			serverConnector.EXPECT().FQDN().Return(fqdn).AnyTimes()
			accessibleConnector.EXPECT().FQDN().Return(altServerAccessible).AnyTimes()
			notAccessibleConnector.EXPECT().FQDN().Return(altServerNotAccessible).AnyTimes()
			serverConnector.EXPECT().Connect(testUser, gomock.Any()).Return(fmt.Errorf("connection error")).AnyTimes()
			notAccessibleConnector.EXPECT().Connect(testUser, gomock.Any()).Return(fmt.Errorf("connection error")).AnyTimes()
			accessibleConnector.EXPECT().Connect(testUser, gomock.Any()).Return(nil).AnyTimes()

			credentialProvider.EXPECT().GetCredentials().Return(credentials, nil).AnyTimes()
			setupServerConnectorMock(accessibleConnector)

			connection := ServerConnectionInfo{
				Server:     fqdn,
				Port:       1022,
				AltAddress: []string{altServerNotAccessible, altServerAccessible},
			}
			apps, err := executor.Discover(context.Background(), connection)

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

func MatchApp(app string, jdkVersion string, springBootVersion string, jarLocation string, instanceCount int) types.GomegaMatcher {
	return PointTo(MatchFields(IgnoreExtras, Fields{
		"AppName": Equal(app),
		"AppType": Equal(string(SpringBootFatJar)),
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
			"JvmMemoryInMB": BeNumerically(">", 0),
			"BindingPorts":  Not(BeEmpty()),
			"OsName":        Equal(osName),
			"OsVersion":     Equal(osVersion),
		})),
		"SpringBootVersion": Equal(springBootVersion),
		"BuildJdkVersion":   MatchVersion(jdkVersion),
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

func setupServerConnectorMock(s *MockServerConnector) {
	s.EXPECT().Close().MinTimes(1)
	s.EXPECT().RunCmd(gomock.Eq(GetLocateJarCmd(mock.SpringBoot1xProcessId, mock.SpringBoot1xJarFile))).Return(mock.SpringBoot1xJarFileLocation, nil).AnyTimes()
	s.EXPECT().RunCmd(gomock.Eq(GetEnvCmd(mock.SpringBoot1xProcessId))).Return(mock.TestEnv, nil).AnyTimes()
	s.EXPECT().RunCmd(gomock.Eq(GetPortsCmd(mock.SpringBoot1xProcessId))).Return(mock.Ports, nil).AnyTimes()
	s.EXPECT().RunCmd(gomock.Eq(GetProcessScanCmd())).Return(mock.SpringBoot1xProcess, nil).AnyTimes()

	s.EXPECT().RunCmd(CmdMatcher(LinuxGetTotalMemoryCmd)).Return(mock.TotalMemory, nil).AnyTimes()
	s.EXPECT().RunCmd(CmdMatcher(LinuxGetJdkVersionCmd)).Return(mock.RuntimeJdkVersion, nil).AnyTimes()
	s.EXPECT().RunCmd(CmdMatcher(LinuxGetDefaultMaxHeapCmd)).Return(mock.DefaultMaxHeapSize, nil).AnyTimes()
	s.EXPECT().RunCmd(CmdMatcher(LinuxSha256Cmd)).Return("", nil).AnyTimes()
	s.EXPECT().RunCmd(CmdMatcher(GetOsName())).Return(osName, nil).AnyTimes()
	s.EXPECT().RunCmd(CmdMatcher(GetOsVersion())).Return(osVersion, nil).AnyTimes()
	s.EXPECT().FQDN().Return(mock.Host).AnyTimes()

	b, err := os.ReadFile(filepath.Join("..", "mock", mock.SpringBoot2xJarFile))
	if err != nil {
		panic(err)
	}
	info, _ := os.Stat(filepath.Join("..", "mock", mock.SpringBoot2xJarFile))
	s.EXPECT().Read(gomock.Eq(mock.SpringBoot2xJarFileLocation)).Return(bytes.NewReader(b), info, nil).AnyTimes()

	b, err = os.ReadFile(filepath.Join("..", "mock", mock.SpringBoot1xJarFile))
	if err != nil {
		panic(err)
	}
	info, _ = os.Stat(filepath.Join("..", "mock", mock.SpringBoot1xJarFile))
	s.EXPECT().Read(gomock.Eq(mock.SpringBoot1xJarFileLocation)).Return(bytes.NewReader(b), info, nil).AnyTimes()

	b, err = os.ReadFile(filepath.Join("..", "mock", mock.ExecutableJarFile))
	if err != nil {
		panic(err)
	}
	info, _ = os.Stat(filepath.Join("..", "mock", mock.ExecutableJarFile))
	s.EXPECT().Read(gomock.Eq(mock.ExecutableJarFileLocation)).Return(bytes.NewReader(b), info, nil).AnyTimes()
}
