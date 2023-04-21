package springboot

// Logging
type Logging struct {
	ConsoleOutput ConsoleOutput `yaml:"console_output"`
	FilePatterns  []string      `yaml:"file_patterns"`
}

// ConsoleOutput
type ConsoleOutput struct {
	Patterns []string `yaml:"patterns"`
	Yamlpath []string `yaml:"yamlpath"`
}

// Static
type Static struct {
	Folder    []string `yaml:"folder"`
	Extension []string `yaml:"extension"`
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
	Parallel    bool `yaml:"parallel"`
	Parallelism int  `yaml:"parallelism"`
}

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

