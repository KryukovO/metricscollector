package server

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/KryukovO/metricscollector/internal/server/config"
	"github.com/KryukovO/metricscollector/internal/server/handlers"
	"github.com/KryukovO/metricscollector/internal/storage"
	"github.com/KryukovO/metricscollector/internal/storage/repository/memstorage"
	"github.com/KryukovO/metricscollector/internal/storage/repository/pgstorage"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	cfg *config.Config
	l   *log.Logger
}

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

func (s *Server) Run() error {
	// Инициализация хранилища
	var (
		repo storage.Repo
		err  error
	)

	retries := []int{0}

	for _, r := range strings.Split(s.cfg.Retries, ",") {
		interval, err := strconv.Atoi(r)
		if err != nil {
			return err
		}

		retries = append(retries, interval)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.StorageTimeout)*time.Second)
	defer cancel()

	if s.cfg.DSN != "" {
		repo, err = pgstorage.NewPgStorage(ctx, s.cfg.DSN, s.cfg.Migrations, retries)
	} else {
		repo, err = memstorage.NewMemStorage(ctx, s.cfg.FileStoragePath, s.cfg.Restore, s.cfg.StoreInterval, retries, s.l)
	}

	if err != nil {
		return err
	}

	stor := storage.NewMetricsStorage(repo, s.cfg.StorageTimeout)
	defer stor.Close()

	// Инициализация сервера
	// NOTE: можно также переопределить e.HTTPErrorHandler, чтобы он не заполнял тело ответа
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	if err := handlers.SetHandlers(e, stor, []byte(s.cfg.Key), s.l); err != nil {
		return err
	}

	// Запуск сервера
	s.l.Infof("Server is running on %s...", s.cfg.HTTPAddress)

	return e.Start(s.cfg.HTTPAddress)
}
