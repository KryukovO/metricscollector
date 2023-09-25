package config

// Параметры конфигурации модуля-сервера.
type Config struct {
	HTTPAddress     string `env:"ADDRESS"`           // Адрес эндпоинта сервера (host:port)
	StoreInterval   uint   `env:"STORE_INTERVAL"`    // Интервал сохранения значения метрик в файл в секундах
	FileStoragePath string `env:"FILE_STORAGE_PATH"` // Полное имя файла, куда сохраняются текущие значения метрик
	Restore         bool   `env:"RESTORE"`           // Признак загрузки значений метрик из файла при запуске сервера
	DSN             string `env:"DATABASE_DSN"`      // Адрес подключения к БД
	Key             string `env:"KEY"`               // Ключ аутентификации

	StoreTimeout    uint   // Таймаут выполнения операций с хранилищем
	ShutdownTimeout uint   // Таймаут для graceful shutdown сервера
	Retries         string // Интервалы попыток соединения с хранилищем через запятую
	Migrations      string // Путь до директории с файлами миграции
}

// Создаёт новый конфиг сервера.
func NewConfig() *Config {
	return &Config{}
}
