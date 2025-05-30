package main

import (
	"context"
	"log"

	"github.com/tezos-delegation-service/cmd/tezos-delegation-job/api/http"
	"github.com/tezos-delegation-service/cmd/tezos-delegation-job/config"
	"github.com/tezos-delegation-service/cmd/tezos-delegation-job/job/poller"
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
	l := logger.Log.WithField("component", "tezos-delegation-job")
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

	tzktAPIAdapter, err := tzktapiadapterfactory.New(cfg.TZKTApiAdapter, metricsClient, l)
	if err != nil {
		l.Fatalf("Failed to create TzKT API factory: %v", err)
	}

	server := http.NewServer(cfg.Server.Port, dbAdapter, metricsClient, l).SetupRoutes()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pollerInstance := poller.New(tzktAPIAdapter, dbAdapter, cfg.TZKTApiAdapter.PollingInterval, metricsClient, l)
	go pollerInstance.Run(ctx)

	if err := server.Start(); err != nil {
		l.Fatalf("Failed to start server: %v", err)
	}

	server.WaitForShutdown(cancel)
}
