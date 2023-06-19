package utils

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"time"
)

func Wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)

	for {
		select {
		case <-timer.C:
			timer.Stop()

			return nil

		case <-ctx.Done():
			return ctx.Err()
		}
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
