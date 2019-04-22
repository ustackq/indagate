package config

// Secret special type for storing secrets.
type Secret string

// Config define indagate general config
type Config struct {
	Global *GlobalConfig
}

// GlobalConfig define common config
type GlobalConfig struct {
	Conn BasicAuth `yaml:"conn"`
}

// BasicAuth define conn config
type BasicAuth struct {
	Username     string `yaml:"user"`
	Paasword     Secret `yaml:"password"`
	PaaswordFile string `yaml:"password_file,omitempty" `
}

func New() *Config {
	config := Config{}
	return &config
}