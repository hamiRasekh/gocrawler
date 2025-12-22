package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"embroidery-designs/internal/api"
	"embroidery-designs/internal/browser"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/crawler"
	"embroidery-designs/internal/proxy"
	"embroidery-designs/internal/service"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
	"embroidery-designs/internal/websocket"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()
	
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	
	// Initialize logger
	if err := utils.InitLogger(cfg.Logging.Level, cfg.Logging.Format); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer utils.Logger.Sync()
	
	logger := utils.GetLogger()
	logger.Info("Starting crawler application")
	
	// Initialize database
	db, err := storage.NewPostgres(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	
	// Initialize repository
	repository := storage.NewRepository(db)
	
	// Initialize proxy manager
	proxyManager := proxy.NewManager(repository, cfg)
	if err := proxyManager.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start proxy manager", zap.Error(err))
	}
	defer proxyManager.Stop()
	
	// Initialize browser manager
	browserManager := browser.NewManager(cfg)
	
	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()
	
	// Initialize crawlers
	apiCrawler := crawler.NewAPICrawler(cfg, repository, proxyManager)
	browserCrawler := crawler.NewBrowserCrawler(cfg, repository, browserManager)
	embroideryCrawler := crawler.NewEmbroideryAPICrawler(cfg, repository, proxyManager)
	
	// Initialize services
	taskService := service.NewTaskService(repository)
	crawlerService := service.NewCrawlerService(repository, apiCrawler, browserCrawler, embroideryCrawler, wsHub)
	
	// Set log callback for embroidery crawler to broadcast logs via WebSocket
	embroideryCrawler.SetLogCallback(func(taskID int64, level, message string) {
		crawlerService.BroadcastLog(taskID, level, message)
	})
	
	// Initialize API handlers
	handlers := api.NewHandlers(taskService, crawlerService, proxyManager, repository, cfg)
	
	// Initialize and start server
	server := api.NewServer(cfg, handlers, wsHub)
	
	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("Server error", zap.Error(err))
		}
	}()
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	<-sigChan
	logger.Info("Shutdown signal received")
	
	// Graceful shutdown
	if err := server.Stop(ctx); err != nil {
		logger.Error("Error during server shutdown", zap.Error(err))
	}
	
	logger.Info("Application stopped")
}

