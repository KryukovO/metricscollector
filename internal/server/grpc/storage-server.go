// Package grpc содержит gRPC-компоненту модуля-сервера.
package grpc

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/KryukovO/metricscollector/api/serverpb"
	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/KryukovO/metricscollector/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	log "github.com/sirupsen/logrus"
)

// ErrStorageIsNil возвращается NewStorageServer, если передано неинициализированное хранилище.
var ErrStorageIsNil = errors.New("storage is nil")

// StorageServer описывает gRPC-сервер, предоставляющий интерфейс доступа к хранилищу метрик.
type StorageServer struct {
	pb.UnimplementedStorageServer

	storage storage.Storage
	l       *log.Logger
}

// NewStorageServer возвращает новый экземляр StorageServer.
func NewStorageServer(s storage.Storage, l *log.Logger) (*StorageServer, error) {
	if s == nil {
		return nil, ErrStorageIsNil
	}

	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	return &StorageServer{
		storage: s,
		l:       lg,
	}, nil
}

// Update выполняет обновление единственной метрики.
func (s *StorageServer) Update(ctx context.Context, req *pb.UpdateRequest) (*emptypb.Empty, error) {
	uuid := ctx.Value("uuid")

	var (
		val interface{}
		err error
	)

	switch req.GetMetric().GetType() {
	case pb.MetricType_COUNTER:
		val = req.GetMetric().GetDelta()
	case pb.MetricType_GAUGE:
		val = float64(req.GetMetric().GetValue())
	default:
		s.l.Debugf("[%s] %s", uuid, metric.ErrWrongMetricType)

		return nil, status.Error(codes.InvalidArgument, metric.ErrWrongMetricType.Error())
	}

	mtrc, err := metric.NewMetrics(req.GetMetric().GetId(), metric.MapGRPCToMetricType[req.GetMetric().GetType()], val)
	if errors.Is(err, metric.ErrWrongMetricName) ||
		errors.Is(err, metric.ErrWrongMetricType) || errors.Is(err, metric.ErrWrongMetricValue) {
		s.l.Debugf("[%s] %s", uuid, err.Error())

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		s.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return nil, status.Error(codes.Internal, err.Error())
	}

	err = s.storage.Update(ctx, &mtrc)
	if errors.Is(err, metric.ErrWrongMetricName) ||
		errors.Is(err, metric.ErrWrongMetricType) || errors.Is(err, metric.ErrWrongMetricValue) {
		s.l.Debugf("[%s] %s", uuid, err.Error())

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		s.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// UpdateMany выполняет обновления набора метрик.
func (s *StorageServer) UpdateMany(ctx context.Context, req *pb.UpdateManyRequest) (*emptypb.Empty, error) {
	uuid := ctx.Value("uuid")

	var (
		val     interface{}
		metrics = make([]metric.Metrics, 0, len(req.GetMetrics()))
		err     error
	)

	for _, mtrc := range req.GetMetrics() {
		switch mtrc.GetType() {
		case pb.MetricType_COUNTER:
			val = mtrc.GetDelta()
		case pb.MetricType_GAUGE:
			val = float64(mtrc.GetValue())
		default:
			s.l.Debugf("[%s] %s", uuid, metric.ErrWrongMetricType)

			return nil, status.Error(codes.InvalidArgument, metric.ErrWrongMetricType.Error())
		}

		m, err := metric.NewMetrics(mtrc.GetId(), metric.MapGRPCToMetricType[mtrc.GetType()], val)
		if errors.Is(err, metric.ErrWrongMetricName) ||
			errors.Is(err, metric.ErrWrongMetricType) || errors.Is(err, metric.ErrWrongMetricValue) {
			s.l.Debugf("[%s] %s", uuid, err.Error())

			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if err != nil {
			s.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

			return nil, status.Error(codes.Internal, err.Error())
		}

		metrics = append(metrics, m)
	}

	err = s.storage.UpdateMany(ctx, metrics)
	if errors.Is(err, metric.ErrWrongMetricName) ||
		errors.Is(err, metric.ErrWrongMetricType) || errors.Is(err, metric.ErrWrongMetricValue) {
		s.l.Debugf("[%s] %s", uuid, err.Error())

		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		s.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// Metric возвращает описание метрики из хранилища.
func (s *StorageServer) Metric(ctx context.Context, req *pb.MetricRequest) (*pb.MetricResponse, error) {
	uuid := ctx.Value("uuid")

	v, err := s.storage.GetValue(ctx, metric.MetricType(req.GetType()), req.GetId())
	if errors.Is(err, metric.ErrWrongMetricType) {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err != nil {
		s.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return nil, status.Error(codes.Internal, err.Error())
	}

	if v.ID == "" {
		msg := fmt.Sprintf("metric with name '%s' not found", req.GetId())

		return nil, status.Error(codes.NotFound, msg)
	}

	resp := &pb.MetricResponse{
		Metric: &pb.MetricDescr{
			Id:   v.ID,
			Type: metric.MapMetricTypeToGRPC[v.MType],
		},
	}

	switch {
	case v.Delta != nil:
		resp.Metric.Delta = *v.Delta
	case v.Value != nil:
		resp.Metric.Value = float32(*v.Value)
	}

	return resp, nil
}

// AllMetrics описание всех метрик из хранилища.
func (s *StorageServer) AllMetrics(ctx context.Context, _ *emptypb.Empty) (*pb.AllMetricsResponse, error) {
	uuid := ctx.Value("uuid")

	values, err := s.storage.GetAll(ctx)
	if err != nil {
		s.l.Errorf("[%s] something went wrong: %s", uuid, err.Error())

		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &pb.AllMetricsResponse{
		Metrics: make([]*pb.MetricDescr, 0, len(values)),
	}

	for _, val := range values {
		mtrc := &pb.MetricDescr{
			Id:   val.ID,
			Type: metric.MapMetricTypeToGRPC[val.MType],
		}

		switch {
		case val.Delta != nil:
			mtrc.Delta = *val.Delta
		case val.Value != nil:
			mtrc.Value = float32(*val.Value)
		}

		resp.Metrics = append(resp.GetMetrics(), mtrc)
	}

	return resp, nil
}
