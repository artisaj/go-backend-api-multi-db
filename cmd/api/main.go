package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-database/internal/application/data"
	"api-database/internal/config"
	"api-database/internal/infrastructure/mongo"
	httpserver "api-database/internal/presentation/http"
	"api-database/internal/telemetry"
)

func main() {
	cfg := config.Load()
	logger := telemetry.NewLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mongoClient := mongo.ConnectOptional(ctx, cfg.Mongo.URI)
	if mongoClient == nil {
		logger.Fatal().Msg("failed to connect to MongoDB (required for datasource configs)")
	}
	defer func() {
		if mongoClient != nil {
			_ = mongoClient.Disconnect(context.Background())
		}
	}()

	dsRepo := mongo.NewDataSourceRepository(mongoClient, cfg.Mongo.DBName)
	queryService := data.NewQueryService(dsRepo)
	metrics := telemetry.NewMetrics(1000)
	dataHandler := httpserver.NewDataHandler(queryService, metrics)

	router := httpserver.NewRouter(cfg, logger, dataHandler, dsRepo, metrics)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("http server failed")
		}
	}()

	logger.Info().Int("port", cfg.Port).Str("env", cfg.Env).Msg("http server started")

	<-ctx.Done()
	logger.Info().Msg("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("graceful shutdown failed")
	} else {
		logger.Info().Msg("server stopped")
	}
}
