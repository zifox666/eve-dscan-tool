package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/zifox666/eve-dscan-tool/internal/config"
	"github.com/zifox666/eve-dscan-tool/internal/service/esi"
	"github.com/zifox666/eve-dscan-tool/internal/service/sde"
	"github.com/zifox666/eve-dscan-tool/internal/store/postgres"
	"gorm.io/gorm"
)

type Server struct {
	config ConfigView
	logger *slog.Logger
	db     *gorm.DB
	sdeDB  *gorm.DB
	redis  *redis.Client
	logs   *postgres.RequestLogRepository
	esi    *esi.Client
	sde    *sde.Service
}

type ConfigView struct {
	AppName    string
	AppVersion string
	AppEnv     string
}

type dependencyCheck struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func NewRouter(cfg config.Config, logger *slog.Logger, db *gorm.DB, sdeDB *gorm.DB, redisClient *redis.Client, esiClient *esi.Client, sdeService *sde.Service) http.Handler {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	server := &Server{
		config: ConfigView{
			AppName:    cfg.AppName,
			AppVersion: cfg.AppVersion,
			AppEnv:     cfg.AppEnv,
		},
		logger: logger,
		db:     db,
		sdeDB:  sdeDB,
		redis:  redisClient,
		logs:   postgres.NewRequestLogRepository(db),
		esi:    esiClient,
		sde:    sdeService,
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(server.requestLogger())

	router.GET("/", server.handleIndex)
	router.GET("/healthz", server.handleHealth)
	router.GET("/readyz", server.handleReady)

	return router
}

func (s *Server) handleIndex(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"app":     s.config.AppName,
		"version": s.config.AppVersion,
		"message": "Gin migration skeleton is running.",
	})
}

func (s *Server) handleHealth(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"app":     s.config.AppName,
		"version": s.config.AppVersion,
		"env":     s.config.AppEnv,
	})
}

func (s *Server) handleReady(ctx *gin.Context) {
	checks := map[string]dependencyCheck{
		"postgres": s.checkPostgres(ctx.Request.Context()),
		"sde":      s.checkSDE(ctx.Request.Context()),
		"redis":    s.checkRedis(ctx.Request.Context()),
	}

	status := http.StatusOK
	overall := "ok"
	for _, check := range checks {
		if check.Status != "ok" {
			status = http.StatusServiceUnavailable
			overall = "degraded"
			break
		}
	}

	ctx.JSON(status, gin.H{
		"status": overall,
		"checks": checks,
	})
}

func (s *Server) checkPostgres(parent context.Context) dependencyCheck {
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()

	sqlDB, err := s.db.DB()
	if err != nil {
		return dependencyCheck{Status: "error", Error: err.Error()}
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return dependencyCheck{Status: "error", Error: err.Error()}
	}

	return dependencyCheck{Status: "ok"}
}

func (s *Server) checkRedis(parent context.Context) dependencyCheck {
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()

	if err := s.redis.Ping(ctx).Err(); err != nil {
		return dependencyCheck{Status: "error", Error: err.Error()}
	}

	return dependencyCheck{Status: "ok"}
}

func (s *Server) checkSDE(parent context.Context) dependencyCheck {
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()

	sqlDB, err := s.sdeDB.DB()
	if err != nil {
		return dependencyCheck{Status: "error", Error: err.Error()}
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return dependencyCheck{Status: "error", Error: err.Error()}
	}

	return dependencyCheck{Status: "ok"}
}

func (s *Server) requestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		startedAt := time.Now()
		ctx.Next()

		s.logger.Info(
			"http request",
			"method", ctx.Request.Method,
			"path", ctx.Request.URL.Path,
			"status", ctx.Writer.Status(),
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"client_ip", ctx.ClientIP(),
		)

		logCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if err := s.logs.Create(logCtx, postgres.CreateRequestLogInput{
			ClientIP:      ctx.ClientIP(),
			RequestPath:   ctx.Request.URL.Path,
			RequestMethod: ctx.Request.Method,
			ProcessTimeMS: time.Since(startedAt).Milliseconds(),
			StatusCode:    ctx.Writer.Status(),
		}); err != nil {
			s.logger.Warn("record request log", "error", err)
		}
	}
}
