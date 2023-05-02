package config

type Config struct {
	HTTPAddress string
}

func New() *Config {
	return &Config{}
}
