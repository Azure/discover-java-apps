package springboot

import (
	_ "embed"
	"gopkg.in/yaml.v3"
	"os"
)

//go:embed config.yml
var defaultConfigYaml string

var YamlCfg = NewYamlConfigOrDie(defaultConfigYaml)

var ConfigPathEnvKey = "CONFIG_PATH"

func NewYamlConfigOrDie(defaultConfigYaml string) YamlConfig {
	yg := YamlConfig{}
	var err error
	var b []byte
	if len(os.Getenv(ConfigPathEnvKey)) > 0 {
		cfgFile := os.Getenv(ConfigPathEnvKey)
		b, err = os.ReadFile(cfgFile)
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(b, &yg)
	} else {
		err = yaml.Unmarshal([]byte(defaultConfigYaml), &yg)
	}

	if err != nil {
		panic(err)
	}

	return yg
}
