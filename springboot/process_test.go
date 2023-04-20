package springboot

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Linux java process", func() {
	var (
		pid      = 1
		uid      = 1
		jar      = "hellospring.jar"
		ctrl     *gomock.Controller
		m        *MockServerConnector
		executor *linuxServerDiscovery
		process  javaProcess
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		m = NewMockServerConnector(ctrl)
		executor = &linuxServerDiscovery{server: m}
		process = javaProcess{
			pid:      pid,
			uid:      uid,
			options:  append(TestJvmOptions, jar),
			javaCmd:  JavaCmd,
			executor: executor,
		}
		m.EXPECT().FQDN().Return("mock_server").AnyTimes()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("Locating jar file", func() {
		When("when jvm options is empty", func() {
			BeforeEach(func() {
				process.options = []string{}
			})

			It("should return error to say cannot locate jar file", func() {
				m.EXPECT().RunCmd(gomock.Any()).MaxTimes(0)
				Expect(process.LocateJarFile()).Error().Should(HaveOccurred())
			})

		})

		When("when jvm options has a relative jar file path", func() {
			BeforeEach(func() {
				process.options = append(TestJvmOptions, "../../../../"+jar)
			})
			It("should return the output folder after sanitized", func() {
				m.EXPECT().RunCmd(gomock.Eq(fmt.Sprintf(LinuxLocateJarCmd, pid, jar))).Return(fmt.Sprintf(" /home/azureuser/%s\n", jar), nil)
				Expect(process.LocateJarFile()).Should(Equal(fmt.Sprintf("/home/azureuser/%s", jar)))
			})
		})

		When("when jvm options has a absolute jar file path", func() {
			BeforeEach(func() {
				process.options = append(TestJvmOptions, "/home/root/"+jar)
			})
			It("should return this absolute jar file path directly", func() {
				Expect(process.LocateJarFile()).Should(Equal("/home/root/" + jar))
			})
		})

		When("error occurred when running cmd", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(gomock.Any()).Return("", fmt.Errorf("test error message"))
				Expect(process.LocateJarFile()).Error().Should(HaveOccurred())
			})

		})
	})

	Context("Get runtime jdk version", func() {
		When("got success output", func() {
			It("should return sanitized version", func() {
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxGetJdkVersionCmd, "java")).Return(RuntimeJdkVersion, nil)
				Expect(process.GetRuntimeJdkVersion()).Should(MatchVersion("11"))
			})
		})
		When("got error", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxGetJdkVersionCmd, "java")).Return("", fmt.Errorf("test error message"))
				Expect(process.GetRuntimeJdkVersion()).Error().Should(HaveOccurred())
			})
		})
	})

	Context("Get jvm options", func() {
		When("jvm options is empty", func() {
			BeforeEach(func() {
				process.options = nil
			})

			It("should return nothing", func() {
				Expect(process.GetJvmOptions()).Should(BeEmpty())
			})
		})

		When("jvm options is not empty", func() {
			It("should return something as expected", func() {
				Expect(process.GetJvmOptions()).Should(And(
					ContainElement("-XX:InitialRAMPercentage=60.0"),
					ContainElement("-DtestOption=abc=def")),
				)
				Expect(process.GetJvmOptions()).ShouldNot(Or(
					ContainElement(JavaCmd),
					ContainElement(JarOption),
					ContainElement(jar),
				))
			})
		})
	})

	Context("Get environments", func() {
		When("got success output after run cmd", func() {
			It("should return the output as kv map", func() {
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxGetEnvCmd, 1)).Return(TestEnv, nil).AnyTimes()
				Expect(process.GetEnvironments()).Should(ContainElement("test_option=test"))

			})
		})
		When("got empty output after run cmd", func() {
			It("should return the output as empty map", func() {
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxGetEnvCmd, 1)).Return("", nil).AnyTimes()
				Expect(process.GetEnvironments()).Should(BeEmpty())
			})
		})
		When("got error after run cmd", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxGetEnvCmd, 1)).Return("", fmt.Errorf("test error message")).AnyTimes()
				Expect(process.GetEnvironments()).Error().Should(HaveOccurred())
			})
		})
	})

	Context("Get jvm heap memory size", func() {
		When("Set -XX:MaxRAMPercentage only in jvm options", func() {
			var percentage float64
			BeforeEach(func() {
				percentage = 60.0
				process.options = []string{
					fmt.Sprintf("-XX:MaxRAMPercentage=%f", percentage),
					jar,
				}
			})

			It("should return the jvm heap memory size based on the percentage of total memory", func() {
				totalMemoryInKb := int64(1000000)
				m.EXPECT().RunCmd(LinuxGetTotalMemoryCmd).Return(fmt.Sprintf("%v\n", totalMemoryInKb), nil)
				Expect(process.GetJvmMemory()).Should(Equal(int64(math.Round(float64(totalMemoryInKb*KiB)*percentage) / 100)))
			})
		})

		When("Set -Xmx and -XX:MaxRAMPercentage in jvm options", func() {
			var size int64
			BeforeEach(func() {
				size = 128
				process.options = []string{
					fmt.Sprintf("-XX:MaxRAMPercentage=%f", 60.0),
					fmt.Sprintf("-Xmx%vm", size),
					jar,
				}
			})

			It("should return the jvm heap memory size in -Xmx which has higher priority", func() {
				m.EXPECT().RunCmd(LinuxGetTotalMemoryCmd).MaxTimes(0)
				Expect(process.GetJvmMemory()).Should(Equal(size * MiB))
			})
		})

		When("Set -Xmx in invalid format", func() {
			var size int64
			BeforeEach(func() {
				size = 128
				process.options = []string{
					fmt.Sprintf("-Xmx%vx", size),
					jar,
				}
			})

			It("should return error", func() {
				Expect(process.GetJvmMemory()).Error().Should(HaveOccurred())
			})
		})

		When("nothing set in jvm options", func() {
			var size int64
			BeforeEach(func() {
				size = int64(100000000)
				process.options = []string{}
			})

			It("should get default max jvm heap memory from vm", func() {
				m.EXPECT().RunCmd(GetDefaultMaxHeap(JavaCmd)).Return(fmt.Sprintf("%v\n", size), nil)
				m.EXPECT().RunCmd(LinuxGetTotalMemoryCmd).MaxTimes(0)
				Expect(process.GetJvmMemory()).Should(Equal(size)) // keep 2 digits precision
			})
		})
	})

	Context("Get ports", func() {
		When("got success output after run cmd", func() {
			It("should return ports in list", func() {
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxGetPortsCmd, pid)).Return(Ports, nil)
				Expect(process.GetPorts()).Should(ContainElements(22, 38193, 8080))
			})
		})
		When("got empty output after run cmd", func() {
			It("should return empty list", func() {
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxGetPortsCmd, pid)).Return(fmt.Sprintf(""), nil)
				Expect(process.GetPorts()).Should(BeEmpty())
			})
		})
		When("got error while run cmd", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(fmt.Sprintf(LinuxGetPortsCmd, pid)).Return("", fmt.Errorf("test error message"))
				Expect(process.GetPorts()).Error().ShouldNot(BeNil())
			})
		})
	})

	Context("Get default max heap size ", func() {
		When("got success output", func() {
			It("should return memory size in kb", func() {
				m.EXPECT().RunCmd(GetDefaultMaxHeap(JavaCmd)).Return(DefaultMaxHeapSize, nil)
				Expect(process.getDefaultMaxHeapSize()).Should(Equal(int64(987654321)))
			})
		})
		When("got empty output", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(GetDefaultMaxHeap(JavaCmd)).Return("  \n", nil)
				Expect(process.getDefaultMaxHeapSize()).Error().Should(HaveOccurred())
			})
		})
		When("got invalid output", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(GetDefaultMaxHeap(JavaCmd)).Return("  abcdefg\n", nil)
				Expect(process.getDefaultMaxHeapSize()).Error().Should(HaveOccurred())
			})
		})
		When("got error", func() {
			It("should return error", func() {
				m.EXPECT().RunCmd(GetDefaultMaxHeap(JavaCmd)).Return("", fmt.Errorf("test error message"))
				Expect(process.getDefaultMaxHeapSize()).Error().Should(HaveOccurred())
			})
		})
	})

})
