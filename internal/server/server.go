// Package server содержит реализацию модуля-сервера.
package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/KryukovO/metricscollector/internal/server/config"
	"github.com/KryukovO/metricscollector/internal/server/http/handlers"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/KryukovO/metricscollector/internal/storage/repository/memstorage"
	"github.com/KryukovO/metricscollector/internal/storage/repository/pgstorage"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// Server - структура сервера.
type Server struct {
	cfg        *config.Config
	httpServer *echo.Echo
	l          *log.Logger
}

// NewServer создаёт новый объект структуры сервера.
func NewServer(cfg *config.Config, l *log.Logger) *Server {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	// Инициализация сервера
	// NOTE: можно также переопределить e.HTTPErrorHandler, чтобы он не заполнял тело ответа
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	return &Server{
		cfg:        cfg,
		httpServer: e,
		l:          lg,
	}
}

// Run инициирует запуск HTTP-сервера и хранилища.
func (s *Server) Run(ctx context.Context) error {
	// Инициализация хранилища
	var (
		repo storage.Repo
		err  error
	)

	retries := []int{0}

	for _, r := range strings.Split(s.cfg.Retries, ",") {
		interval, conertErr := strconv.Atoi(r)
		if conertErr != nil {
			return conertErr
		}

		retries = append(retries, interval)
	}

	sigCtx, sigCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer sigCancel()

	repoCtx, cancel := context.WithTimeout(sigCtx, s.cfg.StoreTimeout.Duration)
	defer cancel()

	s.l.Info("Connecting to the repository...")

	if s.cfg.DSN != "" {
		repo, err = pgstorage.NewPgStorage(repoCtx, s.cfg.DSN, s.cfg.Migrations, retries)
	} else {
		repo, err = memstorage.NewMemStorage(
			repoCtx, s.cfg.FileStoragePath, s.cfg.Restore,
			s.cfg.StoreInterval.Duration, retries, s.l,
		)
	}

	if err != nil {
		return err
	}

	stor := storage.NewMetricsStorage(repo, s.cfg.StoreTimeout.Duration)
	defer func() {
		stor.Close()

		s.l.Info("Repository closed")
	}()

	var ipNet *net.IPNet

	if s.cfg.TrustedSNet != "" {
		_, ipNet, err = net.ParseCIDR(s.cfg.TrustedSNet)
		if err != nil {
			return err
		}
	}

	if err := handlers.SetHandlers(s.httpServer, stor, []byte(s.cfg.Key), s.cfg.PrivateKey, ipNet, s.l); err != nil {
		return err
	}

	g, groupCtx := errgroup.WithContext(ctx)

	// Запуск HTTP-сервера
	g.Go(s.runHTTPServer)

	// Ожидание сигнала завершения
	g.Go(func() error {
		select {
		case <-groupCtx.Done():
			return nil
		case <-sigCtx.Done():
		}

		s.l.Info("Stopping server...")

		shutdownCtx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout.Duration)
		defer cancel()

		s.shutdown(shutdownCtx)

		return nil
	})

	return g.Wait()
}

func (s *Server) runHTTPServer() error {
	s.l.Infof("Run server at %s...", s.cfg.HTTPAddress)

	if err := s.httpServer.Start(s.cfg.HTTPAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

func (s *Server) shutdown(ctx context.Context) {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.l.Errorf("Can't gracefully shutdown server: %s", err.Error())
	} else {
		s.l.Info("Server stopped gracefully")
	}
}
