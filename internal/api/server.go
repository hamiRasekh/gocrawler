package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/utils"
	"embroidery-designs/internal/websocket"
)

type Server struct {
	router   *gin.Engine
	server   *http.Server
	handlers *Handlers
	config   *config.Config
	logger   *zap.Logger
	wsHub    *websocket.Hub
}

func NewServer(cfg *config.Config, handlers *Handlers, wsHub *websocket.Hub) *Server {
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	router.Use(LoggerMiddleware())
	router.Use(RecoveryMiddleware())
	router.Use(CORSMiddleware(cfg.Server.CORSOrigin))

	// Serve Swagger UI and the generated OpenAPI document
	router.Static("/swagger", "./docs/swagger")
	
	SetupRoutes(router, handlers, cfg, handlers.repository, wsHub)
	
	// Serve React static files
	router.Static("/static", "./frontend/dist/assets")
	router.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")
	
	// Serve React app for all non-API routes (SPA routing)
	router.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes
		path := c.Request.URL.Path
		if len(path) < 4 || path[:4] != "/api" {
			c.File("./frontend/dist/index.html")
		}
	})
	
	return &Server{
		router:   router,
		handlers: handlers,
		config:   cfg,
		logger:   utils.GetLogger(),
		wsHub:    wsHub,
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:      router,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server",
		zap.String("host", s.config.Server.Host),
		zap.Int("port", s.config.Server.Port),
	)
	
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}
	
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}
	
	return nil
}

