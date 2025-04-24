package main

import (
	"log"

	"github.com/tezos-delegation-service/cmd/tezos-delegation-api/api/http"
	"github.com/tezos-delegation-service/cmd/tezos-delegation-api/config"
	databaseadapterfactory "github.com/tezos-delegation-service/internal/adapter/database/factory"
	metricsfactory "github.com/tezos-delegation-service/internal/adapter/metrics/factory"
	"github.com/tezos-delegation-service/pkg/logger"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.Setup(&cfg.Logging)
	l := logger.Log.WithField("component", "tezos-delegation-api")
	l.Infof("Service configuration: %+v", cfg)

	metricsClient, err := metricsfactory.New(cfg.Metrics)
	if err != nil {
		l.Fatalf("Failed to create metrics client: %v", err)
	}

	dbAdapter, err := databaseadapterfactory.New(cfg.DatabaseAdapter, metricsClient)
	if err != nil {
		l.Fatalf("Failed to create repository factory: %v", err)
	}
	defer func() {
		if err := dbAdapter.Close(); err != nil {
			l.Errorf("Failed to close repository factory: %v", err)
		}
	}()

	server := http.NewServer(cfg.Server.Port, cfg.Pagination.Limit, dbAdapter, metricsClient, l).SetupRoutes()

	if err := server.Start(); err != nil {
		l.Fatalf("Failed to start server: %v", err)
	}

	server.WaitForShutdown()
}
