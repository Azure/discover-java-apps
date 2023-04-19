package springboot

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
)

var _ = Describe("test config loader", func() {

	When("load from default config", func() {
		var (
			yamlCfg          YamlConfig
			defaultYamlBytes []byte
		)

		BeforeEach(func() {
			defaultYamlBytes, _ = os.ReadFile(filepath.Join("..", "mock", "config.test.yml"))
			yamlCfg = NewYamlConfigOrDie(string(defaultYamlBytes))
		})

		It("should return values from default config", func() {
			Expect(yamlCfg.Server.Connect.Parallel).Should(Equal(false))
			Expect(yamlCfg.Server.Connect.TimeoutSeconds).Should(BeNumerically(">", 0))

			Expect(yamlCfg.Pattern.Cert).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.App).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.Static.Folder).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.Static.Extension).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.Logging.FilePatterns).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.Logging.ConsoleOutput.Patterns).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.Logging.ConsoleOutput.Yamlpath).Should(HaveLen(1))

			Expect(yamlCfg.Env.Denylist).Should(HaveLen(1))
		})
	})

	When("load from env", func() {
		var (
			yamlCfg YamlConfig
		)

		BeforeEach(func() {
			os.Setenv(ConfigPathEnvKey, filepath.Join("..", "mock", "config.test.yml"))
			yamlCfg = NewYamlConfigOrDie("")
		})

		It("should return values from env config", func() {
			Expect(yamlCfg.Server.Connect.Parallel).Should(Equal(false))
			Expect(yamlCfg.Server.Connect.TimeoutSeconds).Should(BeNumerically(">", 0))

			Expect(yamlCfg.Pattern.Cert).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.App).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.Static.Folder).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.Static.Extension).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.Logging.FilePatterns).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.Logging.ConsoleOutput.Patterns).Should(HaveLen(1))
			Expect(yamlCfg.Pattern.Logging.ConsoleOutput.Yamlpath).Should(HaveLen(1))

			Expect(yamlCfg.Env.Denylist).Should(HaveLen(1))
		})
	})
})
