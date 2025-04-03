package main

import (
	"context"
	"log"

	"github.com/tezos-delegation-service/cmd/tezos-delegation-service/api/http"
	"github.com/tezos-delegation-service/cmd/tezos-delegation-service/config"
	"github.com/tezos-delegation-service/cmd/tezos-delegation-service/job/poller"
	databaseadapterfactory "github.com/tezos-delegation-service/internal/adapter/database/factory"
	metricsfactory "github.com/tezos-delegation-service/internal/adapter/metrics/factory"
	tzktapiadapterfactory "github.com/tezos-delegation-service/internal/adapter/tzktapi/factory"
	"github.com/tezos-delegation-service/pkg/logger"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.Setup(&cfg.Logging)
	l := logger.Log.WithField("component", "tezos-delegation-service")
	l.Info("Starting Tezos Delegation API Service...")

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

	tzktAPIAdapter, err := tzktapiadapterfactory.New(cfg.TZKTApiAdapter, metricsClient, l)
	if err != nil {
		l.Fatalf("Failed to create TzKT API factory: %v", err)
	}

	server := http.NewServer(cfg.Server.Port, cfg.Pagination.Limit, dbAdapter, metricsClient, l).SetupRoutes()
	go (poller.New(tzktAPIAdapter, dbAdapter, cfg.TZKTApiAdapter.PollingInterval, metricsClient, l)).Run(context.Background())

	server.WaitForShutdown()

	if err := server.Start(); err != nil {
		l.Fatalf("Failed to start server: %v", err)
	}
}
