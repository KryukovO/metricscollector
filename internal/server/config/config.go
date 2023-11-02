// Package config описывает конфигурацию модуля-сервера.
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/KryukovO/metricscollector/internal/utils"
	"github.com/caarlos0/env"
)

const (
	httpAddress     = "localhost:8080"       // Адрес эндпоинта сервера (host:port) по умолчанию
	storeInterval   = 300 * time.Second      // Интервал сохранения значения метрик в файл в секундах по умолчанию
	fileStoragePath = "/tmp/metrics-db.json" // Полное имя файла, куда сохраняются текущие значения метрик по умолчанию
	restore         = true                   // Признак загрузки значений метрик из файла при запуске сервера по умолчанию
	dsn             = ""                     // Адрес подключения к БД по умолчанию
	key             = ""                     // Ключ аутентификации по умолчанию
	cryptoKey       = "private.key"          // Путь до файла с приватным ключом

	storeTimeout    = 5 * time.Second  // Таймаут выполнения операций с хранилищем по умолчанию
	shutdownTimeout = 10 * time.Second // Таймаут для graceful shutdown сервера по умолчанию
	retries         = "1,3,5"          // Интервалы попыток соединения с хранилищем через запятую по умолчанию
	migrations      = "sql/migrations" // Путь до директории с файлами миграции по умолчанию
)

// ErrPrivateKeyNotFound возвращается, если не был найден публичный ключ шифрования.
var ErrPrivateKeyNotFound = errors.New("private RSA key data not found")

// Config содержит параметры конфигурации модуля-сервера.
type Config struct {
	// HTTPAddress - Адрес эндпоинта сервера (host:port)
	HTTPAddress string `env:"ADDRESS" json:"address"`
	// StoreInterval - Интервал сохранения значения метрик в файл в секундах
	StoreInterval utils.Duration `env:"STORE_INTERVAL" json:"store_interval"`
	// FileStoragePath - Полное имя файла, куда сохраняются текущие значения метрик
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	// Restore - Признак загрузки значений метрик из файла при запуске сервера
	Restore bool `env:"RESTORE" json:"restore"`
	// DSN - Адрес подключения к БД
	DSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// Key - Ключ аутентификации
	Key string `env:"KEY" json:"-"`
	// CryptoKey - Путь до файла с приватным ключом
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`

	// StoreTimeout -Таймаут выполнения операций с хранилищем
	StoreTimeout utils.Duration `json:"-"`
	// ShutdownTimeout - Таймаут для graceful shutdown сервера
	ShutdownTimeout utils.Duration `json:"-"`
	// Retries - Интервалы попыток соединения с хранилищем через запятую
	Retries string `json:"-"`
	// Migrations - Путь до директории с файлами миграции
	Migrations string `json:"-"`
	// PrivateKey - Значение приватного ключа
	PrivateKey rsa.PrivateKey `json:"-"`
}

// NewConfig создаёт новый конфиг сервера.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	var configPath string

	flag.StringVar(&configPath, "c", "", "Configuration file path")
	flag.StringVar(&configPath, "config", "", "Configuration file path")

	flag.StringVar(&cfg.HTTPAddress, "a", httpAddress, "Server endpoint address")
	flag.DurationVar(&cfg.StoreInterval.Duration, "i", storeInterval, "Store interval")
	flag.StringVar(&cfg.FileStoragePath, "f", fileStoragePath, "File storage path")
	flag.BoolVar(&cfg.Restore, "r", restore, "Restore")
	flag.StringVar(&cfg.DSN, "d", dsn, "Data source name")
	flag.StringVar(&cfg.Key, "k", key, "Server key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cryptoKey, "Path to file with private cryptographic key")

	flag.DurationVar(&cfg.StoreTimeout.Duration, "timeout", storeTimeout, "Storage connection timeout")
	flag.DurationVar(&cfg.ShutdownTimeout.Duration, "shutdown", shutdownTimeout, "Graceful shutdown timeout")
	flag.StringVar(&cfg.Retries, "retries", retries, "Server connect retry intervals")
	flag.StringVar(&cfg.Migrations, "migrations", migrations, "Directory of database migration files")

	flag.Parse()

	if configPath != "" {
		err := cfg.parseFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("config file parse error: %w", err)
		}
	}

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

func (cfg *Config) parseFile(path string) error {
	fileConf := &Config{}

	cfgContent, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(cfgContent, fileConf)
	if err != nil {
		return err
	}

	if !utils.IsFlagPassed("a") {
		cfg.HTTPAddress = fileConf.HTTPAddress
	}

	if !utils.IsFlagPassed("r") {
		cfg.Restore = fileConf.Restore
	}

	if !utils.IsFlagPassed("i") {
		cfg.StoreInterval = fileConf.StoreInterval
	}

	if !utils.IsFlagPassed("f") {
		cfg.FileStoragePath = fileConf.FileStoragePath
	}

	if !utils.IsFlagPassed("d") {
		cfg.DSN = fileConf.DSN
	}

	if !utils.IsFlagPassed("crypto-key") {
		cfg.CryptoKey = fileConf.CryptoKey
	}

	return nil
}
