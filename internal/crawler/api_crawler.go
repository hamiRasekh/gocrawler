package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/fingerprint"
	"embroidery-designs/internal/proxy"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
)

type APICrawler struct {
	config        *config.Config
	repository    *storage.Repository
	proxyManager  *proxy.Manager
	rateLimiter   *utils.DomainRateLimiter
	retryConfig   utils.RetryConfig
	logger        *zap.Logger
}

func NewAPICrawler(cfg *config.Config, repo *storage.Repository, proxyMgr *proxy.Manager) *APICrawler {
	return &APICrawler{
		config:       cfg,
		repository:   repo,
		proxyManager: proxyMgr,
		rateLimiter:  utils.NewDomainRateLimiter(cfg.Crawler.RateLimitPerSecond),
		retryConfig: utils.RetryConfig{
			MaxAttempts:       cfg.Crawler.RetryMaxAttempts,
			BackoffMultiplier: cfg.Crawler.RetryBackoffMultiplier,
			InitialDelay:      time.Second,
			MaxDelay:          30 * time.Second,
		},
		logger: utils.GetLogger(),
	}
}

func (ac *APICrawler) Crawl(ctx context.Context, task *storage.Task) error {
	startTime := time.Now()
	
	// Parse task config
	var taskConfig map[string]interface{}
	if task.Config != "" {
		if err := json.Unmarshal([]byte(task.Config), &taskConfig); err != nil {
			ac.logger.Warn("Failed to parse task config, using defaults", zap.Error(err))
		}
	}
	
	// Get proxy if enabled
	var currentProxy *storage.Proxy
	if ac.config.Proxy.Enabled {
		var err error
		currentProxy, err = ac.proxyManager.GetProxy()
		if err != nil {
			ac.logger.Warn("Failed to get proxy, continuing without proxy", zap.Error(err))
		}
	}
	
	// Get HTTP client
	client, err := ac.proxyManager.GetHTTPClient(ctx, currentProxy)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", task.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Apply fingerprint headers
	profile := fingerprint.GenerateProfile()
	fingerprint.ApplyHeaders(req, profile)
	
	// Apply custom headers from config if any
	if headers, ok := taskConfig["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if str, ok := v.(string); ok {
				req.Header.Set(k, str)
			}
		}
	}
	
	// Rate limiting
	if err := ac.rateLimiter.WaitForDomain(ctx, req.URL.Host); err != nil {
		return fmt.Errorf("rate limiter error: %w", err)
	}
	
	// Execute request with retry
	var resp *http.Response
	var proxyUsed *string
	
	if currentProxy != nil {
		proxyStr := fmt.Sprintf("%s://%s:%d", currentProxy.Type, currentProxy.Host, currentProxy.Port)
		proxyUsed = &proxyStr
	}
	
	err = utils.Retry(ctx, ac.retryConfig, func() error {
		var retryErr error
		resp, retryErr = client.Do(req)
		if retryErr != nil {
			// Report proxy failure if using proxy
			if currentProxy != nil {
				ac.proxyManager.ReportProxyFailure(ctx, currentProxy)
				// Try to get new proxy
				currentProxy, _ = ac.proxyManager.GetProxy()
				if currentProxy != nil {
					client, _ = ac.proxyManager.GetHTTPClient(ctx, currentProxy)
				}
			}
			return retryErr
		}
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to execute request after retries: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	
	responseTime := int(time.Since(startTime).Milliseconds())
	
	// Convert headers to JSON
	headersMap := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headersMap[k] = v[0]
		}
	}
	headersJSON, _ := json.Marshal(headersMap)
	
	// Save result
	result := &storage.CrawlResult{
		TaskID:      task.ID,
		URL:         task.URL,
		Method:      req.Method,
		StatusCode:  resp.StatusCode,
		Headers:     string(headersJSON),
		Body:        string(bodyBytes),
		ResponseTime: responseTime,
		ProxyUsed:   proxyUsed,
	}
	
	if err := ac.repository.CreateCrawlResult(ctx, result); err != nil {
		ac.logger.Error("Failed to save crawl result",
			zap.Int64("task_id", task.ID),
			zap.Error(err),
		)
	}
	
	ac.logger.Info("API crawl completed",
		zap.Int64("task_id", task.ID),
		zap.String("url", task.URL),
		zap.Int("status_code", resp.StatusCode),
		zap.Int("response_time_ms", responseTime),
	)
	
	return nil
}

func (ac *APICrawler) Stop(ctx context.Context, taskID int64) error {
	return ac.repository.UpdateTaskStatus(ctx, taskID, storage.TaskStatusStopped)
}

func (ac *APICrawler) Pause(ctx context.Context, taskID int64) error {
	return ac.repository.UpdateTaskStatus(ctx, taskID, storage.TaskStatusPaused)
}

func (ac *APICrawler) Resume(ctx context.Context, taskID int64) error {
	return ac.repository.UpdateTaskStatus(ctx, taskID, storage.TaskStatusRunning)
}

