package config

type Config struct {
	PollInterval   uint   `env:"POLL_INTERVAL"`   // Интервал обновления метрик в секундах
	ReportInterval uint   `env:"REPORT_INTERVAL"` // Интервал отправки метрик в секундах
	ServerAddress  string `env:"ADDRESS"`         // Адрес эндпоинта сервера (host:port)
}

func NewConfig() *Config {
	return &Config{}
}
