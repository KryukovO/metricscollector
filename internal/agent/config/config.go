package config

type Config struct {
	PollInterval   uint   `env:"POLL_INTERVAL"`   // Интервал обновления метрик в секундах
	ReportInterval uint   `env:"REPORT_INTERVAL"` // Интервал отправки метрик в секундах
	ServerAddress  string `env:"ADDRESS"`         // Адрес эндпоинта сервера (host:port)
	Key            string `env:"KEY"`             // Ключ аутентификации

	HTTPTimeout uint   // Таймаут соединения с сервером
	BatchSize   uint   // Количество посылаемых за раз метрик
	Retries     string // Интервалы попыток соединения с сервером через запятую
}

func NewConfig() *Config {
	return &Config{}
}
