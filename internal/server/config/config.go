package config

type Config struct {
	HTTPAddress     string `env:"ADDRESS"`           // Адрес эндпоинта сервера (host:port)
	StoreInterval   uint   `env:"STORE_INTERVAL"`    // Интервал сохранения значения метрик в файл в секундах
	FileStoragePath string `env:"FILE_STORAGE_PATH"` // Полное имя файла, куда сохраняются текущие значения метрик
	Restore         bool   `env:"RESTORE"`           // Признак загрузки значений метрик из файла при запуске сервера
	DSN             string `env:"DATABASE_DSN"`      // Адрес подключения к БД
	StorageTimeout  uint   // Таймаут соединения с хранилищем
	Retries         string // Интервалы попыток соединения с хранилищем через запятую
	Migrations      string // Путь до директории с файлами миграции
}

func NewConfig() *Config {
	return &Config{}
}
