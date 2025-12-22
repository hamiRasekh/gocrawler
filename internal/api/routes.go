package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/websocket"
)

func SetupRoutes(router *gin.Engine, handlers *Handlers, cfg *config.Config, repository *storage.Repository, wsHub *websocket.Hub) {
	api := router.Group("/api/v1")
	
	// Public routes (no auth required)
	api.GET("/health", handlers.Health)
	
	// Auth routes (no auth required)
	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.RegisterAdmin)
		auth.POST("/login", handlers.Login)
		auth.POST("/refresh", handlers.RefreshToken)
		auth.POST("/logout", handlers.Logout)
		auth.POST("/admin-token", handlers.GenerateAdminAPIToken)
	}
	
	// Protected routes (require JWT)
	api.Use(JWTAuthMiddleware(cfg, repository))
	
	api.GET("/stats", handlers.GetStats)
	
	// Tasks
	tasks := api.Group("/tasks")
	{
		tasks.POST("", handlers.CreateTask)
		tasks.GET("", handlers.ListTasks)
		tasks.GET("/:id", handlers.GetTask)
		tasks.PUT("/:id", handlers.UpdateTask)
		tasks.DELETE("/:id", handlers.DeleteTask)
		tasks.POST("/:id/start", handlers.StartTask)
		tasks.POST("/:id/stop", handlers.StopTask)
		tasks.POST("/:id/pause", handlers.PauseTask)
		tasks.GET("/:id/status", handlers.GetTaskStatus)
		tasks.GET("/:id/results", handlers.GetTaskResults)
		tasks.DELETE("/:id/results", handlers.DeleteTaskResults)
	}
	
	// Proxies
	proxies := api.Group("/proxies")
	{
		proxies.GET("", handlers.ListProxies)
		proxies.POST("", handlers.CreateProxy)
		proxies.DELETE("/:id", handlers.DeleteProxy)
		proxies.POST("/test", handlers.TestProxy)
	}
	
	// Products
	products := api.Group("/products")
	{
		products.GET("", handlers.ListProducts)
		products.GET("/stats", handlers.GetProductStats)
		products.GET("/:id", handlers.GetProduct)
		products.GET("/elastic/:elastic_id", handlers.GetProductByElasticID)
		products.DELETE("/:id", handlers.DeleteProduct)
		products.POST("/crawl", handlers.StartEmbroideryCrawl)
		products.GET("/crawl-config", handlers.GetEmbroideryCrawlConfig)
		products.PUT("/crawl-config", handlers.UpdateEmbroideryCrawlConfig)
		products.PATCH("/:id/status", handlers.UpdateProductStatus)
	}
	
	// Auth management (protected)
	authMgmt := api.Group("/auth")
	{
		authMgmt.GET("/me", handlers.GetCurrentUser)
		authMgmt.POST("/tokens", handlers.GenerateToken)
		authMgmt.GET("/tokens", handlers.ListTokens)
		authMgmt.DELETE("/tokens/:id", handlers.DeleteToken)
	}
	
	// WebSocket route (no auth for now, can add later)
	router.GET("/ws/logs", func(c *gin.Context) {
		var taskID *int64
		if taskIDStr := c.Query("task_id"); taskIDStr != "" {
			if id, err := strconv.ParseInt(taskIDStr, 10, 64); err == nil {
				taskID = &id
			}
		}
		websocket.HandleWebSocket(wsHub, c.Writer, c.Request, taskID)
	})
}

