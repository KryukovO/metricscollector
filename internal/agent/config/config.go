// Package config описывает конфигурацию модуля-агента.
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
	serverAddress  = "localhost:8080" // Адрес сервера-хранилища по умолчанию
	reportInterval = 10 * time.Second // Интервал отправки метрик в хранилище в секундах по умолчанию
	pollInterval   = 2 * time.Second  // Интервал сканирования метрик в секундах по умолчанию
	key            = ""               // Значения ключа аутентификации по умолчанию
	rateLimit      = 3                // Количество одновременно исходящих запросов на сервер по умолчанию
	cryptoKey      = "public.key"     // Путь до файла с публичным ключом

	httpTimeout = 5 * time.Second // Таймаут соединения с сервером по умолчанию
	batchSize   = 5               // Количество посылаемых за раз метрик по умолчанию
	retries     = "1,3,5"         // Интервалы попыток соединения с сервером через запятую по умолчанию
)

// ErrPublicKeyNotFound возвращается, если не был найден публичный ключ шифрования.
var ErrPublicKeyNotFound = errors.New("public RSA key data not found")

// Config содержит параметры конфигурации модуля-агента.
type Config struct {
	// PollInterval - Интервал обновления метрик в секундах
	PollInterval utils.Duration `env:"POLL_INTERVAL" json:"poll_interval"`
	// ReportInterval - Интервал отправки метрик в секундах
	ReportInterval utils.Duration `env:"REPORT_INTERVAL" json:"report_interval"`
	// ServerAddress - Адрес эндпоинта сервера (host:port)
	ServerAddress string `env:"ADDRESS" json:"address"`
	// Key - Ключ аутентификации
	Key string `env:"KEY" json:"-"`
	// RateLimit - Количество одновременно исходящих запросов на сервер
	RateLimit uint `env:"RATE_LIMIT" json:"-"`
	// CryptoKey - Путь до файла с публичным ключом
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`

	// HTTPTimeout - Таймаут соединения с сервером
	HTTPTimeout utils.Duration `json:"-"`
	// BatchSize - Количество посылаемых за раз метрик
	BatchSize uint `json:"-"`
	// Retries - Интервалы попыток соединения с сервером через запятую
	Retries string `json:"-"`
	// PublicKey - Значение публичного ключа
	PublicKey rsa.PublicKey `json:"-"`
}

// NewConfig создаёт новый конфиг агента.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	var configPath string

	flag.StringVar(&configPath, "c", "", "Configuration file path")
	flag.StringVar(&configPath, "config", "", "Configuration file path")

	flag.StringVar(&cfg.ServerAddress, "a", serverAddress, "Server endpoint address")
	flag.DurationVar(&cfg.ReportInterval.Duration, "r", reportInterval, "Metric reporting frequency in second")
	flag.DurationVar(&cfg.PollInterval.Duration, "p", pollInterval, "Metric polling frequency in seconds")
	flag.StringVar(&cfg.Key, "k", key, "Server key")
	flag.UintVar(&cfg.RateLimit, "l", rateLimit, "Number of concurrent requests")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cryptoKey, "Path to file with public cryptographic key")

	flag.DurationVar(&cfg.HTTPTimeout.Duration, "timeout", httpTimeout, "Server connection timeout")
	flag.UintVar(&cfg.BatchSize, "batch", batchSize, "Metrics batch size")
	flag.StringVar(&cfg.Retries, "retries", retries, "Server connection retry intervals")

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
		return nil, ErrPublicKeyNotFound
	}

	publicKey, err := x509.ParsePKCS1PublicKey(pkPEM.Bytes)
	if err != nil {
		return nil, err
	}

	cfg.PublicKey = *publicKey

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
		cfg.ServerAddress = fileConf.ServerAddress
	}

	if !utils.IsFlagPassed("r") {
		cfg.ReportInterval = fileConf.ReportInterval
	}

	if !utils.IsFlagPassed("p") {
		cfg.PollInterval = fileConf.PollInterval
	}

	if !utils.IsFlagPassed("crypto-key") {
		cfg.CryptoKey = fileConf.CryptoKey
	}

	return nil
}
