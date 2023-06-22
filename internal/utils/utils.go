package utils

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"time"
)

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

func HashSHA256(src, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)

	_, err := h.Write(src)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
