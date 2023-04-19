package springboot

// Server
type Server struct {
	Connect Connect `yaml:"connect"`
}

// Connect
type Connect struct {
	TimeoutSeconds int  `yaml:"timeoutSeconds"`
	Parallel       bool `yaml:"parallel"`
}

// Pattern
type Pattern struct {
	App     []string `yaml:"app"`
	Logging Logging  `yaml:"logging"`
	Cert    []string `yaml:"cert"`
	Static  Static   `yaml:"static"`
}

// Logging
type Logging struct {
	FilePatterns  []string      `yaml:"file_patterns"`
	ConsoleOutput ConsoleOutput `yaml:"console_output"`
}

// ConsoleOutput
type ConsoleOutput struct {
	Patterns []string `yaml:"patterns"`
	Yamlpath []string `yaml:"yamlpath"`
}

// Static
type Static struct {
	Extension []string `yaml:"extension"`
	Folder    []string `yaml:"folder"`
}

// Env
type Env struct {
	Denylist []string `yaml:"denylist"`
}

// YamlConfig
type YamlConfig struct {
	Server  Server  `yaml:"server"`
	Pattern Pattern `yaml:"pattern"`
	Env     Env     `yaml:"env"`
}

