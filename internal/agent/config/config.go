package config

type Config struct {
	PollInterval   uint   // Интервал обновления метрик в секундах
	ReportInterval uint   // Интервал отправки метрик в секундах
	ServerAddress  string // Адрес эндпоинта сервера (host:port)
}

func New() *Config {
	return &Config{}
}
