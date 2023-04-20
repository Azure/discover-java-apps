package springboot

// YamlConfig
type YamlConfig struct {
	Pattern Pattern `yaml:"pattern"`
	Env     Env     `yaml:"env"`
	Server  Server  `yaml:"server"`
}

// Pattern
type Pattern struct {
	Logging Logging  `yaml:"logging"`
	Cert    []string `yaml:"cert"`
	Static  Static   `yaml:"static"`
	App     []string `yaml:"app"`
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

// Server
type Server struct {
	Connect Connect `yaml:"connect"`
}

// Connect
type Connect struct {
	Parallel       bool `yaml:"parallel"`
	TimeoutSeconds int  `yaml:"timeoutSeconds"`
}

