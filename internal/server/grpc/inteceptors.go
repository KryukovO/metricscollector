package grpc

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Manager предназначен для управления interceptors.
type Manager struct {
	trustedSNet *net.IPNet
	l           *log.Logger
}

// NewManager создаёт новый объект Manager.
func NewManager(trustedSNet *net.IPNet, l *log.Logger) *Manager {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	return &Manager{
		trustedSNet: trustedSNet,
		l:           lg,
	}
}

// LoggingInterceptor - выполняет логгирование входящего gRPC запроса.
func (itc *Manager) LoggingInterceptor(
	ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (interface{}, error) {
	uuid := uuid.New()

	uuidCtx := metadata.AppendToOutgoingContext(ctx, "uuid", uuid.String())

	itc.l.Infof("[%s] received gRPC request: %s", uuid, info.FullMethod)

	ts := time.Now()
	resp, err := handler(uuidCtx, req)
	st, _ := status.FromError(err)

	itc.l.Printf(
		"[%s] query response status: %d; duration: %s",
		uuid, st.Code(), time.Since(ts),
	)

	return resp, err
}

// IPValidationInterceptor - выполняет проверку IP отправителя запроса
// на соответствие доверенной подсети.
func (itc *Manager) IPValidationInterceptor(
	ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (interface{}, error) {
	var ipStr string

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("X-Real-IP")
		if len(values) > 0 {
			ipStr = values[0]
		}
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, status.Error(codes.PermissionDenied, "access is denied by IP")
	}

	if !itc.trustedSNet.Contains(ip) {
		return nil, status.Error(codes.PermissionDenied, "access is denied by IP")
	}

	return handler(ctx, req)
}
