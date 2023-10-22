// Package config описывает конфигурацию модуля-агента.
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
	serverAddress  = "localhost:8080" // Адрес сервера-хранилища по умолчанию
	reportInterval = 10               // Интервал отправки метрик в хранилище в секундах по умолчанию
	pollInterval   = 2                // Интервал сканирования метрик в секундах по умолчанию
	key            = ""               // Значения ключа аутентификации по умолчанию
	rateLimit      = 3                // Количество одновременно исходящих запросов на сервер по умолчанию
	cryptoKey      = "public.key"     // Путь до файла с публичным ключом

	httpTimeout = 5       // Таймаут соединения с сервером по умолчанию
	batchSize   = 5       // Количество посылаемых за раз метрик по умолчанию
	retries     = "1,3,5" // Интервалы попыток соединения с сервером через запятую по умолчанию
)

// ErrPublicKeyNotFound возвращается, если не был найден публичный ключ шифрования.
var ErrPublicKeyNotFound = errors.New("public RSA key data not found")

// Config содержит параметры конфигурации модуля-агента.
type Config struct {
	PollInterval   uint   `env:"POLL_INTERVAL"`   // Интервал обновления метрик в секундах
	ReportInterval uint   `env:"REPORT_INTERVAL"` // Интервал отправки метрик в секундах
	ServerAddress  string `env:"ADDRESS"`         // Адрес эндпоинта сервера (host:port)
	Key            string `env:"KEY"`             // Ключ аутентификации
	RateLimit      uint   `env:"RATE_LIMIT"`      // Количество одновременно исходящих запросов на сервер
	CryptoKey      string `env:"CRYPTO_KEY"`      // Путь до файла с публичным ключом

	HTTPTimeout uint          // Таймаут соединения с сервером
	BatchSize   uint          // Количество посылаемых за раз метрик
	Retries     string        // Интервалы попыток соединения с сервером через запятую
	PublicKey   rsa.PublicKey // Значение публичного ключа
}

// NewConfig создаёт новый конфиг агента.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", serverAddress, "Server endpoint address")
	flag.UintVar(&cfg.ReportInterval, "r", reportInterval, "Metric reporting frequency in second")
	flag.UintVar(&cfg.PollInterval, "p", pollInterval, "Metric polling frequency in seconds")
	flag.StringVar(&cfg.Key, "k", key, "Server key")
	flag.UintVar(&cfg.RateLimit, "l", rateLimit, "Number of concurrent requests")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cryptoKey, "Path to file with public cryptographic key")

	flag.UintVar(&cfg.HTTPTimeout, "timeout", httpTimeout, "Server connection timeout")
	flag.UintVar(&cfg.BatchSize, "batch", batchSize, "Metrics batch size")
	flag.StringVar(&cfg.Retries, "retries", retries, "Server connection retry intervals")

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
		return nil, ErrPublicKeyNotFound
	}

	publicKey, err := x509.ParsePKCS1PublicKey(pkPEM.Bytes)
	if err != nil {
		return nil, err
	}

	cfg.PublicKey = *publicKey

	return cfg, nil
}
