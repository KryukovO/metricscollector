package agent

import (
	"context"
	"testing"
	"time"

	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/mocks"
	"github.com/KryukovO/metricscollector/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend(t *testing.T) {
	var (
		counterVal int64 = 100
		gaugeVal         = 12345.67
		storage          = []metric.Metrics{
			{
				ID:    "PollCount",
				MType: metric.CounterMetric,
				Delta: &counterVal,
			},
			{
				ID:    "RandomValue",
				MType: metric.GaugeMetric,
				Value: &gaugeVal,
			},
		}
	)

	server := mocks.NewMockServer()
	defer server.Close()

	sender, err := NewSender(
		&config.Config{
			Retries:       "1",
			ServerAddress: server.URL,
			Key:           "secret",
			RateLimit:     2,
			HTTPTimeout:   utils.Duration{Duration: 10 * time.Second},
			BatchSize:     1,
		},
		nil,
	)

	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = sender.Send(ctx, storage)
	assert.NoError(t, err)

	<-ctx.Done()
}
