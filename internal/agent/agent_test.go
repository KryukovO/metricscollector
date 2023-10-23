package agent

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/KryukovO/metricscollector/internal/mocks"
	"github.com/KryukovO/metricscollector/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXxx(t *testing.T) {
	server := mocks.NewMockServer()
	defer server.Close()

	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	agent, err := NewAgent(
		&config.Config{
			PollInterval:   utils.Duration{Duration: 1 * time.Second},
			ReportInterval: utils.Duration{Duration: 2 * time.Second},
			Retries:        "1",
			ServerAddress:  server.URL,
			Key:            "secret",
			RateLimit:      2,
			HTTPTimeout:    utils.Duration{Duration: 10 * time.Second},
			BatchSize:      1,
			PublicKey:      privateKey.PublicKey,
		},
		logrus.StandardLogger(),
	)

	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = agent.Run(ctx)
	assert.NoError(t, err)
}
