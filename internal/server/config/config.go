package config

type Config struct {
	HTTPAddress string `env:"ADDRESS"`
}

func New() *Config {
	return &Config{}
}
