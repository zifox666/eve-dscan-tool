package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zifox666/eve-dscan-tool/internal/config"
	"github.com/zifox666/eve-dscan-tool/internal/httpserver"
	"github.com/zifox666/eve-dscan-tool/internal/service/esi"
	"github.com/zifox666/eve-dscan-tool/internal/service/sde"
	"github.com/zifox666/eve-dscan-tool/internal/store/postgres"
	redisstore "github.com/zifox666/eve-dscan-tool/internal/store/redis"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := postgres.Open(cfg)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}

	sqlDB, err := postgres.SQLDB(db)
	if err != nil {
		logger.Error("get postgres sql db", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			logger.Warn("close postgres", "error", err)
		}
	}()

	if err := postgres.AutoMigrate(db); err != nil {
		logger.Error("auto migrate postgres", "error", err)
		os.Exit(1)
	}

	var sdeDBCloser func()
	sdeDB, err := postgres.OpenSDE(cfg)
	if err != nil {
		logger.Error("open sde postgres", "error", err)
		os.Exit(1)
	}
	sdeSQLDB, err := postgres.SQLDB(sdeDB)
	if err != nil {
		logger.Error("get sde sql db", "error", err)
		os.Exit(1)
	}
	sdeDBCloser = func() {
		if err := sdeSQLDB.Close(); err != nil {
			logger.Warn("close sde postgres", "error", err)
		}
	}
	defer sdeDBCloser()

	redisClient := newRedisClient(cfg)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Warn("close redis client", "error", err)
		}
	}()

	jsonCache := redisstore.NewJSONCache(redisClient)
	esiCache := esi.NewCache(jsonCache, cfg.ESICharacterTTL, cfg.ESIAffiliationTTL, cfg.ESIEntityNameTTL)
	esiClient := esi.NewClient(&http.Client{Timeout: cfg.HTTPTimeout}, esiCache, esi.ClientConfig{
		BaseURL:      cfg.ESIBaseURL,
		Datasource:   cfg.ESIDatasource,
		Language:     cfg.ESILanguage,
		MaxRetries:   cfg.ESIMaxRetries,
		RetryDelay:   cfg.ESIRetryDelay,
		IDsBatchSize: cfg.ESIIDsBatchSize,
		BatchSize:    cfg.ESIBatchSize,
	})
	sdeService := sde.NewService(sdeDB)

	router := httpserver.NewRouter(cfg, logger, db, sdeDB, redisClient, esiClient, sdeService)
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting server", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}

func newRedisClient(cfg config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
}
