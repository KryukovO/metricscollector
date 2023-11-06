package sender

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"time"

	pb "github.com/KryukovO/metricscollector/api/serverpb"
	"github.com/KryukovO/metricscollector/internal/agent/config"
	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/utils"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCSender предоставляет функционал взаимодействия с сервером-хранилищем посредством gRPC.
type GRPCSender struct {
	serverAddress string
	rateLimit     uint
	timeout       time.Duration
	batchSize     uint
	retries       []int
	ip            string
	l             *log.Logger
}

// NewGRPCSender создаёт новый объект GRPCSender.
func NewGRPCSender(cfg *config.Config, l *log.Logger) (*GRPCSender, error) {
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

	return &GRPCSender{
		serverAddress: cfg.GRPCAddress,
		rateLimit:     cfg.RateLimit,
		timeout:       cfg.ServerTimeout.Duration,
		batchSize:     cfg.BatchSize,
		retries:       retries,
		ip:            ip.String(),
		l:             lg,
	}, nil
}

// Send инициирует отправку набора метрик в хранилище.
func (snd *GRPCSender) Send(ctx context.Context, storage []metric.Metrics) error {
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
// и инициирует отправку их в хранилище посредством gRPC.
func (snd *GRPCSender) sendTaskWorker(ctx context.Context, id int, tasks <-chan []metric.Metrics) error {
	conn, err := grpc.Dial(snd.serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	defer conn.Close()

	client := pb.NewStorageClient(conn)

	for batch := range tasks {
		select {
		case <-ctx.Done():
			return nil
		default:
			send := func() error {
				sendCtx, cancel := context.WithTimeout(ctx, snd.timeout)
				defer cancel()

				return snd.sendMetrics(sendCtx, client, batch)
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

// sendMetrics выполняет отправку метрик посредством gRPC.
func (snd *GRPCSender) sendMetrics(ctx context.Context, client pb.StorageClient, batch []metric.Metrics) error {
	if client == nil {
		return ErrClientIsNil
	}

	metrics := &pb.UpdateManyRequest{
		Metrics: make([]*pb.MetricDescr, 0, len(batch)),
	}

	for _, mtrc := range batch {
		m := &pb.MetricDescr{
			Id:   mtrc.ID,
			Type: metric.MapMetricTypeToGRPC[mtrc.MType],
		}

		switch {
		case mtrc.Delta != nil:
			m.Delta = *mtrc.Delta
		case mtrc.Value != nil:
			m.Value = float32(*mtrc.Value)
		}

		metrics.Metrics = append(metrics.GetMetrics(), m)
	}

	md := metadata.New(map[string]string{"X-Real-IP": snd.ip})
	grpcCtx := metadata.NewOutgoingContext(ctx, md)

	_, err := client.UpdateMany(grpcCtx, metrics)
	if err != nil {
		st, _ := status.FromError(err)

		return fmt.Errorf("%s: %w", st.Code(), ErrUnexpectedStatus)
	}

	return nil
}
