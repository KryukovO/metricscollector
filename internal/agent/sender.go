package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/utils"
	log "github.com/sirupsen/logrus"
)

type Sender struct {
	serverAddress string
	httpTimeout   time.Duration
	batchSize     uint
	retries       []int
	key           string
	l             *log.Logger
}

func NewSender(cfg *config.Config, l *log.Logger) (*Sender, error) {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	retries := []int{0}

	for _, r := range strings.Split(cfg.Retries, ",") {
		interval, err := strconv.Atoi(r)
		if err != nil {
			return nil, err
		}

		retries = append(retries, interval)
	}

	return &Sender{
		serverAddress: cfg.ServerAddress,
		httpTimeout:   time.Duration(cfg.HTTPTimeout) * time.Second,
		batchSize:     cfg.BatchSize,
		retries:       retries,
		key:           cfg.Key,
		l:             lg,
	}, nil
}

func (snd *Sender) InitMetricSend(storage []metric.Metrics) error {
	if storage == nil {
		return ErrStorageIsNil
	}

	mtrcs := make([]metric.Metrics, 0, snd.batchSize)

	sendBatch := func(mtrcs []metric.Metrics) error {
		var (
			err    error
			client http.Client
		)

		send := func() error {
			ctx, cancel := context.WithTimeout(context.Background(), snd.httpTimeout)
			defer cancel()

			return snd.sendMetrics(ctx, &client, mtrcs)
		}

		for _, t := range snd.retries {
			err = utils.Wait(context.Background(), time.Duration(t)*time.Second)
			if err != nil {
				return err
			}

			err = send()
			if err == nil || !errors.Is(err, syscall.ECONNREFUSED) {
				break
			}
		}

		if err != nil {
			snd.l.Errorf("error sending metric values: %s", err.Error())
		} else {
			snd.l.Debugf("metrics sent: %d", len(mtrcs))
		}

		return nil
	}

	for _, mtrc := range storage {
		mtrcs = append(mtrcs, mtrc)

		if len(mtrcs) == int(snd.batchSize) {
			err := sendBatch(mtrcs)
			if err != nil {
				return err
			}

			mtrcs = make([]metric.Metrics, 0, snd.batchSize)
		}
	}

	if len(mtrcs) != 0 {
		err := sendBatch(mtrcs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (snd *Sender) sendMetrics(ctx context.Context, client *http.Client, batch []metric.Metrics) error {
	if client == nil {
		return ErrClientIsNil
	}

	url := fmt.Sprintf("http://%s/updates/", snd.serverAddress)

	body, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	gz := gzip.NewWriter(buf)
	if _, err = gz.Write(body); err != nil {
		return err
	}

	gz.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	if snd.key != "" {
		hash, err := utils.HashSHA256(body, []byte(snd.key))
		if err != nil {
			return err
		}

		req.Header.Set("HashSHA256", hex.EncodeToString(hash))
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	_, err = io.Copy(io.Discard, resp.Body)

	defer resp.Body.Close()

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %w", resp.Status, ErrUnexpectedStatus)
	}

	return nil
}
