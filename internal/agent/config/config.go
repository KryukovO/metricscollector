package config

// Параметры конфигурации модуля-агента.
type Config struct {
	PollInterval   uint   `env:"POLL_INTERVAL"`   // Интервал обновления метрик в секундах
	ReportInterval uint   `env:"REPORT_INTERVAL"` // Интервал отправки метрик в секундах
	ServerAddress  string `env:"ADDRESS"`         // Адрес эндпоинта сервера (host:port)
	Key            string `env:"KEY"`             // Ключ аутентификации
	RateLimit      uint   `env:"RATE_LIMIT"`      // Количество одновременно исходящих запросов на сервер

	HTTPTimeout uint   // Таймаут соединения с сервером
	BatchSize   uint   // Количество посылаемых за раз метрик
	Retries     string // Интервалы попыток соединения с сервером через запятую
}

// Создаёт новый конфиг агента.
func NewConfig() *Config {
	return &Config{}
}
