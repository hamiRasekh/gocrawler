package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"embroidery-designs/internal/browser"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
)

type BrowserCrawler struct {
	config      *config.Config
	repository  *storage.Repository
	browserMgr  *browser.Manager
	logger      *zap.Logger
}

func NewBrowserCrawler(cfg *config.Config, repo *storage.Repository, browserMgr *browser.Manager) *BrowserCrawler {
	return &BrowserCrawler{
		config:     cfg,
		repository: repo,
		browserMgr: browserMgr,
		logger:     utils.GetLogger(),
	}
}

func (bc *BrowserCrawler) Crawl(ctx context.Context, task *storage.Task) error {
	startTime := time.Now()
	
	// Parse task config
	var taskConfig map[string]interface{}
	if task.Config != "" {
		if err := json.Unmarshal([]byte(task.Config), &taskConfig); err != nil {
			bc.logger.Warn("Failed to parse task config", zap.Error(err))
		}
	}
	
	// Generate browser profile
	profile := bc.browserMgr.GenerateProfile()
	bc.browserMgr.SetProfile(profile)
	
	// Create browser context
	browserCtx, cancel, err := bc.browserMgr.CreateContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to create browser context: %w", err)
	}
	defer cancel()
	
	// Apply stealth
	if err := browser.ApplyStealth(browserCtx, profile); err != nil {
		bc.logger.Warn("Failed to apply stealth", zap.Error(err))
	}
	
	// Navigate to URL
	launcher := bc.browserMgr.GetLauncher()
	if err := launcher.Navigate(browserCtx, task.URL); err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	
	// Wait a bit for page to load
	time.Sleep(2 * time.Second)
	
	// Get page content
	html, err := launcher.GetPageContent(browserCtx)
	if err != nil {
		return fmt.Errorf("failed to get page content: %w", err)
	}
	
	responseTime := int(time.Since(startTime).Milliseconds())
	
	// Extract data if needed (can be customized based on task config)
	if extractScript, ok := taskConfig["extract_script"].(string); ok && extractScript != "" {
		_, err := launcher.ExecuteScript(browserCtx, extractScript)
		if err != nil {
			bc.logger.Warn("Failed to execute extract script", zap.Error(err))
		}
	}
	
	// Prepare headers (simulated from browser)
	headersMap := map[string]string{
		"User-Agent":      profile.UserAgent,
		"Accept-Language": profile.AcceptLanguage,
		"Accept":          profile.Accept,
	}
	headersJSON, _ := json.Marshal(headersMap)
	
	// Save result
	result := &storage.CrawlResult{
		TaskID:      task.ID,
		URL:         task.URL,
		Method:      "GET",
		StatusCode:  200, // Browser navigation is always 200 if successful
		Headers:     string(headersJSON),
		Body:        html,
		ResponseTime: responseTime,
		ProxyUsed:   nil, // Browser handles proxy internally if configured
	}
	
	if err := bc.repository.CreateCrawlResult(ctx, result); err != nil {
		bc.logger.Error("Failed to save crawl result",
			zap.Int64("task_id", task.ID),
			zap.Error(err),
		)
	}
	
	bc.logger.Info("Browser crawl completed",
		zap.Int64("task_id", task.ID),
		zap.String("url", task.URL),
		zap.Int("response_time_ms", responseTime),
	)
	
	return nil
}

func (bc *BrowserCrawler) Stop(ctx context.Context, taskID int64) error {
	return bc.repository.UpdateTaskStatus(ctx, taskID, storage.TaskStatusStopped)
}

func (bc *BrowserCrawler) Pause(ctx context.Context, taskID int64) error {
	return bc.repository.UpdateTaskStatus(ctx, taskID, storage.TaskStatusPaused)
}

func (bc *BrowserCrawler) Resume(ctx context.Context, taskID int64) error {
	return bc.repository.UpdateTaskStatus(ctx, taskID, storage.TaskStatusRunning)
}

