package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
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

// HTTPSender предоставляет функционал взаимодействия с сервером-хранилищем посредством HTTP.
type HTTPSender struct {
	serverAddress string
	rateLimit     uint
	timeout       time.Duration
	batchSize     uint
	retries       []int
	key           string
	publicKey     *rsa.PublicKey
	ip            string
	l             *log.Logger
}

// NewHTTPSender создаёт новый объект HTTPSender.
func NewHTTPSender(cfg *config.Config, l *log.Logger) (*HTTPSender, error) {
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

	ip, err := utils.LocalIP()
	if err != nil {
		return nil, err
	}

	return &HTTPSender{
		serverAddress: cfg.HTTPAddress,
		rateLimit:     cfg.RateLimit,
		timeout:       cfg.ServerTimeout.Duration,
		batchSize:     cfg.BatchSize,
		retries:       retries,
		key:           cfg.Key,
		publicKey:     cfg.PublicKey,
		ip:            ip.String(),
		l:             lg,
	}, nil
}

// Send инициирует отправку набора метрик в хранилище.
func (snd *HTTPSender) Send(ctx context.Context, storage []metric.Metrics) error {
	if storage == nil {
		return ErrStorageIsNil
	}

	g, ctx := errgroup.WithContext(ctx)
	tasks := generateSendTasks(ctx, storage, snd.rateLimit, snd.batchSize)

	for w := 1; w <= int(snd.rateLimit); w++ {
		id := w

		g.Go(func() error {
			return snd.sendTaskWorker(ctx, id, tasks)
		})
	}

	return g.Wait()
}

// sendTaskWorker выполняет сканирование канала на наличие в нем сообщений, содержащих метрики,
// и инициирует отправку их в хранилище посредством HTTP.
func (snd *HTTPSender) sendTaskWorker(ctx context.Context, id int, tasks <-chan []metric.Metrics) error {
	var (
		err    error
		client http.Client
	)

	for batch := range tasks {
		select {
		case <-ctx.Done():
			return nil
		default:
			send := func() error {
				sendCtx, cancel := context.WithTimeout(ctx, snd.timeout)
				defer cancel()

				return snd.sendMetrics(sendCtx, &client, batch)
			}

			for _, t := range snd.retries {
				err = utils.Wait(ctx, time.Duration(t)*time.Second)
				if err != nil {
					break
				}

				err = send()
				if err == nil || !errors.Is(err, syscall.ECONNREFUSED) {
					break
				}
			}

			if err != nil {
				if errors.Is(err, ErrClientIsNil) {
					return err
				}

				snd.l.Errorf("[worker %d] error sending metric values: %s", id, err.Error())
			} else {
				snd.l.Debugf("[worker %d] metrics sent: %d", id, len(batch))
			}
		}
	}

	return nil
}

// sendMetrics выполняет отправку метрик посредством HTTP.
func (snd *HTTPSender) sendMetrics(ctx context.Context, client *http.Client, batch []metric.Metrics) error {
	if client == nil {
		return ErrClientIsNil
	}

	url := fmt.Sprintf("%s/updates/", snd.serverAddress)

	if !strings.HasPrefix(url, "http://") {
		url = fmt.Sprintf("http://%s", url)
	}

	body, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	if snd.publicKey != nil {
		body, err = rsa.EncryptOAEP(sha256.New(), rand.Reader, snd.publicKey, body, nil)
		if err != nil {
			return err
		}
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
	req.Header.Set("X-Real-IP", snd.ip)

	if snd.key != "" {
		hash, hashErr := utils.HashSHA256(body, []byte(snd.key))
		if hashErr != nil {
			return hashErr
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
