package core

import (
	_ "embed"
	"os"

	gookitconfig "github.com/gookit/config/v2"
	gookityaml "github.com/gookit/config/v2/yamlv3"
)

var defaultConfigYaml string

var config = NewDefaultConfig()
var ExportedConfig = config

type Config interface {
	GetConfigStringByKey(key string) string
	GetConfigIntByKey(key string) int
	GetConfigBoolByKey(key string) bool
	GetConfigListByKey(key string) []string
}

type defaultConfig struct {
	cfg *gookitconfig.Config
}

func NewDefaultConfig() Config {
	c := &defaultConfig{}
	c.init()
	return c
}

func (config *defaultConfig) init() {
	cfg := gookitconfig.New("config")
	cfg.AddDriver(gookityaml.Driver)
	var cfgFile string
	var err error
	if len(os.Getenv("CONFIG_PATH")) > 0 {
		cfgFile = os.Getenv("CONFIG_PATH")
		err = cfg.LoadFiles(cfgFile)
	} else {
		err = cfg.LoadStrings(gookitconfig.Yaml, defaultConfigYaml)
	}

	if err != nil {
		panic(err)
	}
	config.cfg = cfg
}

func (config *defaultConfig) GetConfigStringByKey(key string) string {
	return config.cfg.String(key)
}

func (config *defaultConfig) GetConfigIntByKey(key string) int {
	return config.cfg.Int(key, 0)
}

func (config *defaultConfig) GetConfigBoolByKey(key string) bool {
	return config.cfg.Bool(key, false)
}

func (config *defaultConfig) GetConfigListByKey(key string) []string {
	return config.cfg.Strings(key)
}
