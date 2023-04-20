package springboot

// YamlConfig
type YamlConfig struct {
	Server  Server  `yaml:"server"`
	Pattern Pattern `yaml:"pattern"`
	Env     Env     `yaml:"env"`
}

// Server
type Server struct {
	Connect Connect `yaml:"connect"`
}

// Connect
type Connect struct {
	Parallel       bool `yaml:"parallel"`
	TimeoutSeconds int  `yaml:"timeoutSeconds"`
}

// Pattern
type Pattern struct {
	Static  Static   `yaml:"static"`
	App     []string `yaml:"app"`
	Logging Logging  `yaml:"logging"`
	Cert    []string `yaml:"cert"`
}

// Static
type Static struct {
	Folder    []string `yaml:"folder"`
	Extension []string `yaml:"extension"`
}

// Logging
type Logging struct {
	FilePatterns  []string      `yaml:"file_patterns"`
	ConsoleOutput ConsoleOutput `yaml:"console_output"`
}

// ConsoleOutput
type ConsoleOutput struct {
	Yamlpath []string `yaml:"yamlpath"`
	Patterns []string `yaml:"patterns"`
}

// Env
type Env struct {
	Denylist []string `yaml:"denylist"`
}

