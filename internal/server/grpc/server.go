// Package grpc содержит gRPC-компоненту модуля-сервера.
package grpc

import (
	"context"
	"errors"

	pb "github.com/KryukovO/metricscollector/api/serverpb"
	"github.com/KryukovO/metricscollector/internal/storage"
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
	return &emptypb.Empty{}, nil
}

// UpdateMany выполняет обновления набора метрик.
func (s *StorageServer) UpdateMany(ctx context.Context, req *pb.UpdateManyRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// Metric возвращает описание метрики из хранилища.
func (s *StorageServer) Metric(ctx context.Context, req *pb.MetricRequest) (*pb.MetricResponse, error) {
	return &pb.MetricResponse{}, nil
}

// AllMetrics описание всех метрик из хранилища.
func (s *StorageServer) AllMetrics(ctx context.Context, _ *emptypb.Empty) (*pb.AllMetricsResponse, error) {
	return &pb.AllMetricsResponse{}, nil
}
