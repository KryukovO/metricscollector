// Package utils содержит различные функции общего назначения.
package utils

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"flag"
	"time"
)

// Wait предназначена для выполнения ожидания в течение d, или пока не будет прерван контекст.
func Wait(ctx context.Context, d time.Duration) error {
	if d == 0 {
		return ctx.Err()
	}

	ticker := time.NewTicker(d)

	select {
	case <-ticker.C:
		ticker.Stop()

		return nil

	case <-ctx.Done():
		return ctx.Err()
	}
}

// HashSHA256 производить вычисление SHA256 хеша.
func HashSHA256(src, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)

	_, err := h.Write(src)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

// IsFlagPassed проверяет, был ли указан флаг запуска.
func IsFlagPassed(name string) bool {
	found := false

	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})

	return found
}
