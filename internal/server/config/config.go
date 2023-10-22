// Package config описывает конфигурацию модуля-сервера.
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env"
)

const (
	httpAddress     = "localhost:8080"       // Адрес эндпоинта сервера (host:port) по умолчанию
	storeInterval   = 300                    // Интервал сохранения значения метрик в файл в секундах по умолчанию
	fileStoragePath = "/tmp/metrics-db.json" // Полное имя файла, куда сохраняются текущие значения метрик по умолчанию
	restore         = true                   // Признак загрузки значений метрик из файла при запуске сервера по умолчанию
	dsn             = ""                     // Адрес подключения к БД по умолчанию
	key             = ""                     // Ключ аутентификации по умолчанию
	cryptoKey       = "private.key"          // Путь до файла с приватным ключом

	storeTimeout    = 5                // Таймаут выполнения операций с хранилищем по умолчанию
	shutdownTimeout = 10               // Таймаут для graceful shutdown сервера по умолчанию
	retries         = "1,3,5"          // Интервалы попыток соединения с хранилищем через запятую по умолчанию
	migrations      = "sql/migrations" // Путь до директории с файлами миграции по умолчанию
)

// ErrPrivateKeyNotFound возвращается, если не был найден публичный ключ шифрования.
var ErrPrivateKeyNotFound = errors.New("private RSA key data not found")

// Config содержит параметры конфигурации модуля-сервера.
type Config struct {
	HTTPAddress     string `env:"ADDRESS"`           // Адрес эндпоинта сервера (host:port)
	StoreInterval   uint   `env:"STORE_INTERVAL"`    // Интервал сохранения значения метрик в файл в секундах
	FileStoragePath string `env:"FILE_STORAGE_PATH"` // Полное имя файла, куда сохраняются текущие значения метрик
	Restore         bool   `env:"RESTORE"`           // Признак загрузки значений метрик из файла при запуске сервера
	DSN             string `env:"DATABASE_DSN"`      // Адрес подключения к БД
	Key             string `env:"KEY"`               // Ключ аутентификации
	CryptoKey       string `env:"CRYPTO_KEY"`        // Путь до файла с приватным ключом

	StoreTimeout    uint           // Таймаут выполнения операций с хранилищем
	ShutdownTimeout uint           // Таймаут для graceful shutdown сервера
	Retries         string         // Интервалы попыток соединения с хранилищем через запятую
	Migrations      string         // Путь до директории с файлами миграции
	PrivateKey      rsa.PrivateKey // Значение приватного ключа
}

// NewConfig создаёт новый конфиг сервера.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.HTTPAddress, "a", httpAddress, "Server endpoint address")
	flag.UintVar(&cfg.StoreInterval, "i", storeInterval, "Store interval")
	flag.StringVar(&cfg.FileStoragePath, "f", fileStoragePath, "File storage path")
	flag.BoolVar(&cfg.Restore, "r", restore, "Restore")
	flag.StringVar(&cfg.DSN, "d", dsn, "Data source name")
	flag.StringVar(&cfg.Key, "k", key, "Server key")

	flag.UintVar(&cfg.StoreTimeout, "timeout", storeTimeout, "Storage connection timeout")
	flag.UintVar(&cfg.ShutdownTimeout, "shutdown", shutdownTimeout, "Graceful shutdown timeout")
	flag.StringVar(&cfg.Retries, "retries", retries, "Server connect retry intervals")
	flag.StringVar(&cfg.Migrations, "migrations", migrations, "Directory of database migration files")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cryptoKey, "Path to file with private cryptographic key")

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("env parsing error: %w", err)
	}

	content, err := os.ReadFile(cfg.CryptoKey)
	if err != nil {
		return nil, err
	}

	pkPEM, _ := pem.Decode(content)
	if pkPEM == nil {
		return nil, ErrPrivateKeyNotFound
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(pkPEM.Bytes)
	if err != nil {
		return nil, err
	}

	cfg.PrivateKey = *privateKey

	return cfg, nil
}
