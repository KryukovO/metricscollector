package utils

import (
	"context"
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
