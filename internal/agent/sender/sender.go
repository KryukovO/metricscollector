package sender

import (
	"context"
	"errors"

	"github.com/KryukovO/metricscollector/internal/metric"
)

var (
	// ErrStorageIsNil возвращается sender.Send, если был передан неинициализированный слайс метрик.
	ErrStorageIsNil = errors.New("metrics buf is nil")
	// ErrClientIsNil возвращается sender.sendMetrics, если был передан
	// неинициализированный HTTP-клиент или соединение gRPC.
	ErrClientIsNil = errors.New("HTTP client is nil")
	// ErrUnexpectedStatus возвращается sender.sendMetrics, если сервером был возвращен статус отличный от 200 OK.
	ErrUnexpectedStatus = errors.New("unexpected response status")
)

type Sender interface {
	Send(ctx context.Context, storage []metric.Metrics) error
}

// generateSendTasks разбивает набор метрик на батчи определенного размера.
// Передаёт батчи через возвращаемый канал.
func generateSendTasks(
	ctx context.Context, storage []metric.Metrics,
	rateLimit, batchSize uint,
) chan []metric.Metrics {
	outCh := make(chan []metric.Metrics, rateLimit)

	go func() {
		defer close(outCh)

		batch := make([]metric.Metrics, 0, batchSize)

		for _, mtrc := range storage {
			batch = append(batch, mtrc)

			if len(batch) == int(batchSize) {
				select {
				case <-ctx.Done():
					return
				case outCh <- batch:
				}

				batch = make([]metric.Metrics, 0, batchSize)
			}
		}

		if len(batch) != 0 {
			select {
			case <-ctx.Done():
				return
			case outCh <- batch:
			}
		}
	}()

	return outCh
}
