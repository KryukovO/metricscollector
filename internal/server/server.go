// Package server содержит реализацию модуля-сервера.
package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/KryukovO/metricscollector/internal/server/config"
	"github.com/KryukovO/metricscollector/internal/server/handlers"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/KryukovO/metricscollector/internal/storage/repository/memstorage"
	"github.com/KryukovO/metricscollector/internal/storage/repository/pgstorage"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

// Server - структура сервера.
type Server struct {
	cfg *config.Config
	l   *log.Logger
}

// NewServer создаёт новый объект структуры сервера.
func NewServer(cfg *config.Config, l *log.Logger) *Server {
	lg := log.StandardLogger()
	if l != nil {
		lg = l
	}

	return &Server{
		cfg: cfg,
		l:   lg,
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

	sigCtx, sigCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer sigCancel()

	repoCtx, cancel := context.WithTimeout(sigCtx, time.Duration(s.cfg.StoreTimeout)*time.Second)
	defer cancel()

	s.l.Info("Connecting to the repository...")

	if s.cfg.DSN != "" {
		repo, err = pgstorage.NewPgStorage(repoCtx, s.cfg.DSN, s.cfg.Migrations, retries)
	} else {
		repo, err = memstorage.NewMemStorage(repoCtx, s.cfg.FileStoragePath, s.cfg.Restore, s.cfg.StoreInterval, retries, s.l)
	}

	if err != nil {
		return err
	}

	stor := storage.NewMetricsStorage(repo, s.cfg.StoreTimeout)
	defer func() {
		stor.Close()

		s.l.Info("Repository closed")
	}()

	// Инициализация сервера
	// NOTE: можно также переопределить e.HTTPErrorHandler, чтобы он не заполнял тело ответа
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	if err := handlers.SetHandlers(e, stor, []byte(s.cfg.Key), s.l); err != nil {
		return err
	}

	g, groupCtx := errgroup.WithContext(ctx)

	// Запуск сервера
	g.Go(func() error {
		s.l.Infof("Run server at %s...", s.cfg.HTTPAddress)

		if err := e.Start(s.cfg.HTTPAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	// Ожидание сигнала завершения
	g.Go(func() error {
		select {
		case <-groupCtx.Done():
			return nil
		case <-sigCtx.Done():
		}

		s.l.Info("Stopping server...")

		shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(s.cfg.ShutdownTimeout)*time.Second)
		defer cancel()

		if err := e.Shutdown(shutdownCtx); err != nil {
			s.l.Errorf("Can't gracefully shutdown server: %s", err.Error())
		} else {
			s.l.Info("Server stopped gracefully")
		}

		return nil
	})

	return g.Wait()
}
