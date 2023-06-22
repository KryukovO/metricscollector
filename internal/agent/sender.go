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
	"golang.org/x/sync/errgroup"
)

type Sender struct {
	serverAddress string
	rateLimit     uint
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
		rateLimit:     cfg.RateLimit,
		httpTimeout:   time.Duration(cfg.HTTPTimeout) * time.Second,
		batchSize:     cfg.BatchSize,
		retries:       retries,
		key:           cfg.Key,
		l:             lg,
	}, nil
}

func (snd *Sender) Send(storage []metric.Metrics) error {
	if storage == nil {
		return ErrStorageIsNil
	}

	done := make(chan struct{})
	defer close(done)

	tasks := snd.generateSendTasks(done, storage)

	g := new(errgroup.Group)

	for w := 1; w <= int(snd.rateLimit); w++ {
		id := w

		g.Go(func() error {
			return snd.sendTaskWorker(id, tasks)
		})
	}

	return g.Wait()
}

func (snd *Sender) generateSendTasks(done <-chan struct{}, storage []metric.Metrics) chan []metric.Metrics {
	out := make(chan []metric.Metrics, snd.rateLimit)

	go func() {
		defer close(out)

		batch := make([]metric.Metrics, 0, snd.batchSize)

		for _, mtrc := range storage {
			select {
			case <-done:
				return
			default:
				batch = append(batch, mtrc)

				if len(batch) == int(snd.batchSize) {
					out <- batch

					batch = make([]metric.Metrics, 0, snd.batchSize)
				}
			}
		}

		if len(batch) != 0 {
			out <- batch
		}
	}()

	return out
}

func (snd *Sender) sendTaskWorker(id int, tasks <-chan []metric.Metrics) error {
	var (
		err    error
		client http.Client
	)

	for batch := range tasks {
		send := func() error {
			ctx, cancel := context.WithTimeout(context.Background(), snd.httpTimeout)
			defer cancel()

			return snd.sendMetrics(ctx, &client, batch)
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
			snd.l.Errorf("[worker %d] error sending metric values: %s", id, err.Error())
		} else {
			snd.l.Debugf("[worker %d] metrics sent: %d", id, len(batch))
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
