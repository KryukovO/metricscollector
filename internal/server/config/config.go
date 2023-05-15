package config

type Config struct {
	HTTPAddress string `env:"ADDRESS"`
}

func NewConfig() *Config {
	return &Config{}
}
