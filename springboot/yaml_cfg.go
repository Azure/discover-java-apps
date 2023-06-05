package springboot

// Pattern
type Pattern struct {
	Cert    []string `yaml:"cert"`
	Static  Static   `yaml:"static"`
	App     []string `yaml:"app"`
	Logging Logging  `yaml:"logging"`
}

// Static
type Static struct {
	Extension []string `yaml:"extension"`
	Folder    []string `yaml:"folder"`
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

// Server
type Server struct {
	Connect Connect `yaml:"connect"`
}

// Connect
type Connect struct {
	Parallel    bool `yaml:"parallel"`
	Parallelism int  `yaml:"parallelism"`
}
