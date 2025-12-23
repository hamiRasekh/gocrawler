package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"embroidery-designs/internal/auth"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/crawler"
	"embroidery-designs/internal/proxy"
	"embroidery-designs/internal/service"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	taskService    *service.TaskService
	crawlerService *service.CrawlerService
	proxyManager   *proxy.Manager
	repository     *storage.Repository
	config         *config.Config
	logger         *zap.Logger
}

func NewHandlers(
	taskService *service.TaskService,
	crawlerService *service.CrawlerService,
	proxyManager *proxy.Manager,
	repository *storage.Repository,
	cfg *config.Config,
) *Handlers {
	return &Handlers{
		taskService:    taskService,
		crawlerService: crawlerService,
		proxyManager:   proxyManager,
		repository:     repository,
		config:         cfg,
		logger:         utils.GetLogger(),
	}
}

// Task handlers
func (h *Handlers) CreateTask(c *gin.Context) {
	var req service.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	task, err := h.taskService.CreateTask(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

func (h *Handlers) ListTasks(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	tasks, err := h.taskService.ListTasks(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list tasks", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":  tasks,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *Handlers) GetTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handlers) UpdateTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var req service.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	task, err := h.taskService.UpdateTask(c.Request.Context(), id, &req)
	if err != nil {
		h.logger.Error("Failed to update task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handlers) DeleteTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.taskService.DeleteTask(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

func (h *Handlers) StartTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.crawlerService.StartTask(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to start task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task started"})
}

func (h *Handlers) StopTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.crawlerService.StopTask(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to stop task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task stopped"})
}

func (h *Handlers) PauseTask(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.crawlerService.PauseTask(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to pause task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pause task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task paused"})
}

func (h *Handlers) GetTaskStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.taskService.GetTask(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id": task.ID,
		"status":  task.Status,
	})
}

func (h *Handlers) GetTaskResults(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	results, err := h.repository.GetCrawlResults(c.Request.Context(), id, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get results", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *Handlers) DeleteTaskResults(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.repository.DeleteCrawlResults(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete results", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Results deleted"})
}

// Proxy handlers
func (h *Handlers) ListProxies(c *gin.Context) {
	proxies, err := h.repository.ListProxies(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to list proxies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list proxies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"proxies": proxies})
}

func (h *Handlers) CreateProxy(c *gin.Context) {
	var proxy storage.Proxy
	if err := c.ShouldBindJSON(&proxy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy configuration"})
		return
	}

	if err := h.repository.CreateProxy(c.Request.Context(), &proxy); err != nil {
		h.logger.Error("Failed to create proxy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy"})
		return
	}

	// Reload proxy pool
	_ = h.proxyManager.ReloadProxies(c.Request.Context())

	c.JSON(http.StatusCreated, proxy)
}

func (h *Handlers) DeleteProxy(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy ID"})
		return
	}

	if err := h.repository.DeleteProxy(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete proxy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete proxy"})
		return
	}

	// Reload proxy pool
	_ = h.proxyManager.ReloadProxies(c.Request.Context())

	c.JSON(http.StatusOK, gin.H{"message": "Proxy deleted"})
}

func (h *Handlers) TestProxy(c *gin.Context) {
	var proxy storage.Proxy
	if err := c.ShouldBindJSON(&proxy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy configuration"})
		return
	}

	healthy, err := h.proxyManager.TestProxy(c.Request.Context(), &proxy)
	if err != nil {
		h.logger.Error("Failed to test proxy", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test proxy"})
		return
	}

	if proxy.ID != 0 {
		if err := h.repository.UpdateProxyHealth(c.Request.Context(), proxy.ID, healthy); err != nil {
			h.logger.Warn("Failed to update proxy health after manual test", zap.Error(err), zap.Int64("proxy_id", proxy.ID))
		} else {
			_ = h.proxyManager.ReloadProxies(c.Request.Context())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"healthy": healthy,
		"proxy":   proxy,
	})
}

// Health and stats
func (h *Handlers) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

func (h *Handlers) GetStats(c *gin.Context) {
	// Get task stats
	tasks, _ := h.taskService.ListTasks(c.Request.Context(), 1000, 0)

	statusCounts := make(map[string]int)
	for _, task := range tasks {
		statusCounts[string(task.Status)]++
	}

	// Get proxy stats
	proxies, _ := h.repository.ListProxies(c.Request.Context())
	activeProxies := 0
	for _, p := range proxies {
		if p.IsActive {
			activeProxies++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": gin.H{
			"total":     len(tasks),
			"by_status": statusCounts,
		},
		"proxies": gin.H{
			"total":  len(proxies),
			"active": activeProxies,
		},
	})
}

// Product handlers
func (h *Handlers) ListProducts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	// Build filters
	filters := make(map[string]interface{})
	if brand := c.Query("brand"); brand != "" {
		filters["brand"] = brand
	}
	if catalog := c.Query("catalog"); catalog != "" {
		filters["catalog"] = catalog
	}
	if inStock := c.Query("in_stock"); inStock != "" {
		filters["in_stock"] = inStock == "true"
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	var statusValues []string
	if statusParam := c.Query("status"); statusParam != "" {
		statusValues = append(statusValues, strings.Split(statusParam, ",")...)
	}
	for _, value := range c.QueryArray("status") {
		statusValues = append(statusValues, value)
	}
	for _, value := range c.QueryArray("statuses") {
		statusValues = append(statusValues, value)
	}

	if len(statusValues) > 0 {
		var statuses []storage.ProductStatus
		for _, raw := range statusValues {
			status := storage.ProductStatus(strings.ToLower(strings.TrimSpace(raw)))
			if status.IsValid() {
				statuses = append(statuses, status)
			}
		}
		if len(statuses) > 0 {
			filters["statuses"] = statuses
		}
	}

	products, total, err := h.repository.ListProducts(c.Request.Context(), limit, offset, filters)
	if err != nil {
		h.logger.Error("Failed to list products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *Handlers) GetProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	product, err := h.repository.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handlers) GetProductByElasticID(c *gin.Context) {
	elasticID := c.Param("elastic_id")
	if elasticID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid elastic ID"})
		return
	}

	product, err := h.repository.GetProductByElasticID(c.Request.Context(), elasticID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handlers) GetProductStats(c *gin.Context) {
	stats, err := h.repository.GetProductStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get product stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handlers) DeleteProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	if err := h.repository.DeleteProduct(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}

func (h *Handlers) UpdateProductStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req UpdateProductStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	status := storage.ProductStatus(strings.ToLower(strings.TrimSpace(req.Status)))
	if !status.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	if err := h.repository.UpdateProductStatus(c.Request.Context(), id, status); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		h.logger.Error("Failed to update product status", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product status"})
		return
	}

	product, err := h.repository.GetProduct(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *Handlers) StartEmbroideryCrawl(c *gin.Context) {
	// Create a task for embroidery API crawl
	task := &storage.Task{
		Name:   "Embroidery Designs API Crawl",
		URL:    "https://www.embroiderydesigns.com/es/prdsrch",
		Type:   "api",
		Config: `{"crawler_type": "embroidery_api"}`,
	}

	if err := h.repository.CreateTask(c.Request.Context(), task); err != nil {
		h.logger.Error("Failed to create task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Start the task
	if err := h.crawlerService.StartTask(c.Request.Context(), task.ID); err != nil {
		h.logger.Error("Failed to start task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start task"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Embroidery crawl started",
		"task_id": task.ID,
	})
}

func (h *Handlers) GetEmbroideryCrawlConfig(c *gin.Context) {
	overrides, updatedAt, err := h.repository.GetEmbroideryPayloadOverrides(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to load embroidery payload overrides", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load crawl config"})
		return
	}

	if overrides == nil {
		overrides = map[string]interface{}{}
	}

	pageSize := h.config.Embroidery.PageSize
	if pageSize == 0 {
		pageSize = 120
	}

	defaultPayload := crawler.BuildEmbroideryPayload(0, pageSize, nil)
	effectivePayload := crawler.BuildEmbroideryPayload(0, pageSize, overrides)

	response := gin.H{
		"payload_overrides": overrides,
		"default_payload":   defaultPayload,
		"effective_payload": effectivePayload,
	}

	if updatedAt != nil && !updatedAt.IsZero() {
		response["updated_at"] = updatedAt
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) UpdateEmbroideryCrawlConfig(c *gin.Context) {
	var req UpdateEmbroideryConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid configuration format"})
		return
	}

	sanitized := sanitizePayloadOverrides(req.PayloadOverrides)

	if err := h.repository.UpdateEmbroideryPayloadOverrides(c.Request.Context(), sanitized); err != nil {
		h.logger.Error("Failed to persist embroidery payload overrides", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update crawl config"})
		return
	}

	// Return updated config snapshot
	overrides, updatedAt, err := h.repository.GetEmbroideryPayloadOverrides(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to reload embroidery payload overrides", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated crawl config"})
		return
	}

	pageSize := h.config.Embroidery.PageSize
	if pageSize == 0 {
		pageSize = 120
	}

	defaultPayload := crawler.BuildEmbroideryPayload(0, pageSize, nil)
	effectivePayload := crawler.BuildEmbroideryPayload(0, pageSize, overrides)

	response := gin.H{
		"payload_overrides": overrides,
		"default_payload":   defaultPayload,
		"effective_payload": effectivePayload,
	}

	if updatedAt != nil && !updatedAt.IsZero() {
		response["updated_at"] = updatedAt
	}

	c.JSON(http.StatusOK, response)
}

// Auth handlers
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterAdminRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AdminTokenRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	TokenName string `json:"token_name"`
}

type GenerateTokenRequest struct {
	TokenName string `json:"token_name" binding:"required"`
	ExpiresAt string `json:"expires_at" binding:"required"` // Accept as string, parse manually
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateProductStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdateEmbroideryConfigRequest struct {
	PayloadOverrides map[string]interface{} `json:"payload_overrides" binding:"required"`
}

func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get user from database
	user, err := h.repository.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(h.config, user.ID, user.Username)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		h.logger.Error("Failed to generate refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshExpiresAt := time.Now().Add(h.config.Auth.RefreshExpiration)
	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()

	refreshRecord := &storage.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: refreshExpiresAt,
	}

	if userAgent != "" {
		refreshRecord.UserAgent = &userAgent
	}

	if ipAddress != "" {
		refreshRecord.IPAddress = &ipAddress
	}

	if err := h.repository.CreateRefreshToken(c.Request.Context(), refreshRecord); err != nil {
		h.logger.Error("Failed to persist refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":              token,
		"refresh_token":      refreshToken,
		"expires_in":         int(h.config.Auth.JWTExpiration.Seconds()),
		"refresh_expires_at": refreshExpiresAt,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

func (h *Handlers) RegisterAdmin(c *gin.Context) {
	var req RegisterAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	username := strings.TrimSpace(req.Username)
	if len(username) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username must be at least 3 characters"})
		return
	}

	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
		return
	}

	if existing, err := h.repository.GetUserByUsername(c.Request.Context(), username); err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		h.logger.Error("Failed to check existing user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	user := &storage.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	if err := h.repository.CreateUser(c.Request.Context(), user); err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"created_at": user.CreatedAt,
	})
}

func (h *Handlers) GenerateAdminAPIToken(c *gin.Context) {
	var req AdminTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	user, err := h.repository.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenName := strings.TrimSpace(req.TokenName)
	if tokenName == "" {
		tokenName = fmt.Sprintf("%s-admin-token-%d", user.Username, time.Now().Unix())
	}

	lifetime := h.config.Auth.AdminTokenLifetime
	if lifetime <= 0 {
		lifetime = 365 * 24 * time.Hour
	}
	expiresAt := time.Now().Add(lifetime)

	tokenString, err := auth.GenerateAPIToken(
		h.config,
		user.ID,
		user.Username,
		tokenName,
		expiresAt,
	)
	if err != nil {
		h.logger.Error("Failed to generate admin API token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	apiToken := &storage.APIToken{
		UserID:    user.ID,
		TokenName: tokenName,
		TokenHash: hashToken(tokenString),
		ExpiresAt: expiresAt,
	}

	if err := h.repository.CreateAPIToken(c.Request.Context(), apiToken); err != nil {
		h.logger.Error("Failed to persist admin API token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":      tokenString,
		"token_name": tokenName,
		"expires_at": expiresAt,
		"expires_in": int(expiresAt.Sub(time.Now()).Seconds()),
	})
}

func (h *Handlers) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	tokenHash := hashToken(req.RefreshToken)

	refreshToken, err := h.repository.GetRefreshTokenByHash(c.Request.Context(), tokenHash)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	user, err := h.repository.GetUserByID(c.Request.Context(), refreshToken.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	if err := h.repository.RevokeRefreshToken(c.Request.Context(), refreshToken.ID); err != nil {
		h.logger.Warn("Failed to revoke refresh token", zap.Error(err))
	}

	accessToken, err := auth.GenerateToken(h.config, user.ID, user.Username)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	newRefreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		h.logger.Error("Failed to generate refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	expiresAt := time.Now().Add(h.config.Auth.RefreshExpiration)
	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()

	refreshRecord := &storage.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(newRefreshToken),
		ExpiresAt: expiresAt,
	}

	if userAgent != "" {
		refreshRecord.UserAgent = &userAgent
	}

	if ipAddress != "" {
		refreshRecord.IPAddress = &ipAddress
	}

	if err := h.repository.CreateRefreshToken(c.Request.Context(), refreshRecord); err != nil {
		h.logger.Error("Failed to persist refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":              accessToken,
		"refresh_token":      newRefreshToken,
		"expires_in":         int(h.config.Auth.JWTExpiration.Seconds()),
		"refresh_expires_at": expiresAt,
	})
}

func (h *Handlers) Logout(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	tokenHash := hashToken(req.RefreshToken)
	refreshToken, err := h.repository.GetRefreshTokenByHash(c.Request.Context(), tokenHash)
	if err != nil {
		// Token already expired or invalid; treat as success
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
		return
	}

	if err := h.repository.RevokeRefreshToken(c.Request.Context(), refreshToken.ID); err != nil {
		h.logger.Warn("Failed to revoke refresh token on logout", zap.Error(err))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (h *Handlers) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	username, _ := c.Get("username")

	c.JSON(http.StatusOK, gin.H{
		"id":       userID,
		"username": username,
	})
}

func (h *Handlers) GenerateToken(c *gin.Context) {
	var req GenerateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Parse expires_at from datetime-local format (YYYY-MM-DDTHH:mm)
	// Try multiple formats to be flexible
	var expiresAt time.Time
	var parseErr error

	// Try RFC3339 format first (with seconds and timezone)
	expiresAt, parseErr = time.Parse(time.RFC3339, req.ExpiresAt)
	if parseErr != nil {
		// Try datetime-local format (YYYY-MM-DDTHH:mm)
		expiresAt, parseErr = time.Parse("2006-01-02T15:04", req.ExpiresAt)
		if parseErr != nil {
			// Try with seconds
			expiresAt, parseErr = time.Parse("2006-01-02T15:04:05", req.ExpiresAt)
			if parseErr != nil {
				h.logger.Error("Failed to parse expires_at", zap.String("value", req.ExpiresAt), zap.Error(parseErr))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expires_at format. Expected: YYYY-MM-DDTHH:mm"})
				return
			}
		}
	}

	// Validate that expiration is in the future
	if expiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "expires_at must be in the future"})
		return
	}

	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")

	// Generate JWT token
	tokenString, err := auth.GenerateAPIToken(
		h.config,
		userID.(int64),
		username.(string),
		req.TokenName,
		expiresAt,
	)
	if err != nil {
		h.logger.Error("Failed to generate API token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Hash token for storage
	tokenHash := hashToken(tokenString)

	// Store token in database
	apiToken := &storage.APIToken{
		UserID:    userID.(int64),
		TokenName: req.TokenName,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	if err := h.repository.CreateAPIToken(c.Request.Context(), apiToken); err != nil {
		h.logger.Error("Failed to save API token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": tokenString,
		"token_info": gin.H{
			"id":         apiToken.ID,
			"token_name": apiToken.TokenName,
			"expires_at": apiToken.ExpiresAt,
		},
	})
}

func (h *Handlers) ListTokens(c *gin.Context) {
	userID, _ := c.Get("user_id")

	tokens, err := h.repository.ListAPITokens(c.Request.Context(), userID.(int64))
	if err != nil {
		h.logger.Error("Failed to list tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

func (h *Handlers) DeleteToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token ID"})
		return
	}

	if err := h.repository.DeleteAPIToken(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token deleted"})
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func sanitizePayloadOverrides(input map[string]interface{}) map[string]interface{} {
	if input == nil {
		return map[string]interface{}{}
	}

	copied := deepCopyMap(input)
	delete(copied, "from")
	delete(copied, "size")
	return copied
}

func deepCopyMap(src map[string]interface{}) map[string]interface{} {
	if src == nil {
		return map[string]interface{}{}
	}

	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		switch typed := v.(type) {
		case map[string]interface{}:
			dst[k] = deepCopyMap(typed)
		case []interface{}:
			dst[k] = deepCopySlice(typed)
		default:
			dst[k] = typed
		}
	}
	return dst
}

func deepCopySlice(src []interface{}) []interface{} {
	dst := make([]interface{}, len(src))
	for i, v := range src {
		switch typed := v.(type) {
		case map[string]interface{}:
			dst[i] = deepCopyMap(typed)
		case []interface{}:
			dst[i] = deepCopySlice(typed)
		default:
			dst[i] = typed
		}
	}
	return dst
}
