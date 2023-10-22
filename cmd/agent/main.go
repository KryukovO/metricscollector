// Модуль-агент предназначен для сбора метрик из различных источников
// и передачи их в хранилище.
package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/KryukovO/metricscollector/internal/agent"
	"github.com/KryukovO/metricscollector/internal/agent/config"

	log "github.com/sirupsen/logrus"

	_ "net/http/pprof"
)

var (
	// buildVersion представляет собой хранилище значения ldflag - версия сборки.
	buildVersion = "N/A"
	// buildDate представляет собой хранилище значения ldflag - дата сборки.
	buildDate = "N/A"
	// buildCommit представляет собой хранилище значения ldflag - комментарий к сборке.
	buildCommit = "N/A"
)

func main() {
	fmt.Printf(
		"Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		buildVersion, buildDate, buildCommit,
	)

	l := log.New()
	l.SetLevel(log.DebugLevel)
	l.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 Z07:00",
	})

	cfg, err := config.NewConfig()
	if err != nil {
		l.Fatalf("config init error: %v. Exit(1)", err)
	}

	go func() {
		if listenErr := http.ListenAndServe("0.0.0.0:8081", nil); !errors.Is(listenErr, http.ErrServerClosed) {
			l.Errorf("pprof running error: %v", listenErr)
		}
	}()

	agnt, err := agent.NewAgent(cfg, l)
	if err != nil {
		l.Fatalf("agent initialization error: %v. Exit(1)", err)
	}

	if err := agnt.Run(); err != nil {
		l.Fatalf("agent running error: %v. Exit(1)", err)
	}
}
