package weblogic

import (
	_ "embed"
)

//go:embed config.yml
var defaultConfigYaml string

var ConfigPathEnvKey = "CONFIG_PATH"
