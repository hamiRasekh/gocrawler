package crawler

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"embroidery-designs/internal/config"
	"embroidery-designs/internal/fingerprint"
	"embroidery-designs/internal/proxy"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
	"github.com/andybalholm/brotli"
	"go.uber.org/zap"
)

type LogCallback func(taskID int64, level, message string)

type CrawlStats struct {
	TotalProcessed int64
	SuccessCount   int64
	ErrorCount     int64
	Errors         []string
	StartTime      time.Time
	LastUpdate     time.Time
}

type EmbroideryAPICrawler struct {
	config           *config.Config
	repository       *storage.Repository
	proxyManager     *proxy.Manager
	rateLimiter      *utils.DomainRateLimiter
	retryConfig      utils.RetryConfig
	logger           *zap.Logger
	baseURL          string
	apiHeaders       map[string]string
	cookies          string
	logCallback      LogCallback
	periodicMonitors map[int64]context.CancelFunc
	mu               sync.RWMutex
}

func (eac *EmbroideryAPICrawler) SetLogCallback(callback LogCallback) {
	eac.logCallback = callback
}

type ElasticsearchResponse struct {
	Took     int64 `json:"took"`
	TimedOut bool  `json:"timed_out"`
	Hits     struct {
		Total struct {
			Value    int64  `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []struct {
			ID     string                 `json:"_id"`
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func NewEmbroideryAPICrawler(
	cfg *config.Config,
	repo *storage.Repository,
	proxyMgr *proxy.Manager,
) *EmbroideryAPICrawler {
	eac := &EmbroideryAPICrawler{
		config:           cfg,
		repository:       repo,
		proxyManager:     proxyMgr,
		rateLimiter:      utils.NewDomainRateLimiter(cfg.Crawler.RateLimitPerSecond),
		periodicMonitors: make(map[int64]context.CancelFunc),
		retryConfig: utils.RetryConfig{
			MaxAttempts:       cfg.Crawler.RetryMaxAttempts,
			BackoffMultiplier: cfg.Crawler.RetryBackoffMultiplier,
			InitialDelay:      time.Second,
			MaxDelay:          30 * time.Second,
		},
		logger: utils.GetLogger(),
	}

	// Load from config with fallback to defaults
	embCfg := cfg.Embroidery
	eac.baseURL = embCfg.BaseURL
	if eac.baseURL == "" {
		eac.baseURL = "https://www.embroiderydesigns.com/es/prdsrch"
	}

	eac.cookies = embCfg.Cookies
	if eac.cookies == "" {
		// Fallback to default cookies if not provided
		eac.cookies = ".AspNetCore.Antiforgery.TvID2vd87ec=CfDJ8KTVZ637icVPj_pEImxQ47SfvvQDCb4TrQjWAz_ZEFIIi2CL8qEjO7-trnGom56mEmndxk1vDdjCyjQmEPLQQVFX4j3g0n-GlX16zNOqvV1APqRh6UHMD8v_Ht1xDblkYSi2ETDxVbmQMUCZ5GP2Yzs; __cflb=0H28vLEqq6KG1NM4B5fRPAsPJ6QTtcuTzpCMR9YQvPE; CountryCode=US; CookieConsent=1; ThirdPartyConsent=1"
	}

	// API-specific headers (will be merged with fingerprint headers)
	authToken := embCfg.AuthToken
	if authToken == "" {
		authToken = "Basic ZWxhc3RpY1JlYWRPbmx5OnpnOzlTcSFXPnc1O1FoLDJ0eVU="
	}

	eac.apiHeaders = map[string]string{
		"accept":         "application/json, text/plain, */*",
		"authorization":  authToken,
		"content-type":   "application/json;charset=UTF-8",
		"origin":         "https://www.embroiderydesigns.com",
		"referer":        "https://www.embroiderydesigns.com/stockdesign/productlistings",
		"sec-fetch-dest": "empty",
		"sec-fetch-mode": "cors",
		"sec-fetch-site": "same-origin",
	}

	return eac
}

func (eac *EmbroideryAPICrawler) CrawlAll(ctx context.Context, task *storage.Task) error {
	eac.logger.Info("Starting embroidery API crawl", zap.Int64("task_id", task.ID))
	if eac.logCallback != nil {
		eac.logCallback(task.ID, "info", "Starting embroidery API crawl")
	}

	stats := &CrawlStats{
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Errors:     make([]string, 0),
	}

	payloadOverrides := eac.loadPayloadOverrides(ctx)

	// Parse task config for resume support
	var taskConfig map[string]interface{}
	from := 0
	if task.Config != "" {
		if err := json.Unmarshal([]byte(task.Config), &taskConfig); err == nil {
			if lastFrom, ok := taskConfig["last_from"].(float64); ok {
				from = int(lastFrom)
				logMsg := fmt.Sprintf("Resuming from position: %d", from)
				eac.logger.Info(logMsg, zap.Int64("task_id", task.ID))
				if eac.logCallback != nil {
					eac.logCallback(task.ID, "info", logMsg)
				}
			}
		}
	}

	pageSize := eac.config.Embroidery.PageSize
	if pageSize == 0 {
		pageSize = 120
	}

	totalProcessed := from
	var totalAvailable int64 = 0
	startTime := time.Now()
	const batchSize = 50

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			eac.logger.Info("Crawl cancelled", zap.Int64("task_id", task.ID))
			return ctx.Err()
		default:
		}

		// Update task config with current position for resume support
		if taskConfig == nil {
			taskConfig = make(map[string]interface{})
		}
		taskConfig["last_from"] = from
		if configBytes, err := json.Marshal(taskConfig); err == nil {
			// Update task config in database (non-blocking)
			go func() {
				_ = eac.repository.UpdateTaskConfig(ctx, task.ID, string(configBytes))
			}()
		}

		// Create request payload
		payload := eac.createPayload(from, pageSize, payloadOverrides)
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}

		// Make request with fingerprint
		resp, err := eac.makeRequest(ctx, payloadBytes)
		if err != nil {
			eac.logger.Error("Failed to make request",
				zap.Int64("task_id", task.ID),
				zap.Int("from", from),
				zap.Error(err),
			)
			stats.ErrorCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("Request failed at from=%d: %v", from, err))
			
			// Retry with exponential backoff
			time.Sleep(time.Second * 2)
			continue
		}

		// Parse response
		var esResp ElasticsearchResponse
		if err := json.Unmarshal(resp, &esResp); err != nil {
			eac.logger.Error("Failed to parse response",
				zap.Int64("task_id", task.ID),
				zap.Error(err),
			)
			stats.ErrorCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("Parse error at from=%d: %v", from, err))
			continue
		}

		// Set total available on first request
		if totalAvailable == 0 {
			totalAvailable = esResp.Hits.Total.Value
			logMsg := fmt.Sprintf("Total products available: %d", totalAvailable)
			eac.logger.Info(logMsg, zap.Int64("task_id", task.ID))
			if eac.logCallback != nil {
				eac.logCallback(task.ID, "info", logMsg)
			}
		}

		// Process products
		productsCount := len(esResp.Hits.Hits)
		if productsCount == 0 {
			eac.logger.Info("No more products to process", zap.Int64("task_id", task.ID))
			break
		}

		// Save products in batches
		batch := make([]map[string]interface{}, 0, batchSize)
		batchIDs := make([]string, 0, batchSize)

		for i, hit := range esResp.Hits.Hits {
			batch = append(batch, hit.Source)
			batchIDs = append(batchIDs, hit.ID)

			// Save batch when it's full or at the end
			if len(batch) >= batchSize || i == len(esResp.Hits.Hits)-1 {
				successCount, errCount := eac.saveProductsBatch(ctx, batch, batchIDs, stats)
				stats.SuccessCount += successCount
				stats.ErrorCount += errCount
				
				// Clear batch
				batch = make([]map[string]interface{}, 0, batchSize)
				batchIDs = make([]string, 0, batchSize)
			}
		}

		totalProcessed += productsCount
		stats.TotalProcessed = int64(totalProcessed)
		stats.LastUpdate = time.Now()

		// Calculate progress
		remaining := int(totalAvailable) - totalProcessed
		elapsed := time.Since(startTime)
		var estimatedTimeRemaining time.Duration
		if totalProcessed > 0 {
			avgTimePerItem := elapsed / time.Duration(totalProcessed)
			estimatedTimeRemaining = avgTimePerItem * time.Duration(remaining)
		}

		logMsg := fmt.Sprintf("Page %d: %d products processed (Total: %d/%d Remaining: %d) - Estimated time remaining: %v",
			from/pageSize+1, productsCount, totalProcessed, totalAvailable, remaining, estimatedTimeRemaining.Round(time.Second))

		eac.logger.Info("Processed page",
			zap.Int64("task_id", task.ID),
			zap.Int("from", from),
			zap.Int("count", productsCount),
			zap.Int("total_processed", totalProcessed),
			zap.Int64("total_available", totalAvailable),
			zap.Int("remaining", remaining),
			zap.Duration("estimated_time_remaining", estimatedTimeRemaining),
		)
		if eac.logCallback != nil {
			eac.logCallback(task.ID, "info", logMsg)
		}

		// Check if we've processed all products
		if totalProcessed >= int(totalAvailable) || productsCount < pageSize {
			break
		}

		// Move to next page
		from += pageSize

		// Check context before continuing
		select {
		case <-ctx.Done():
			eac.logger.Info("Crawl cancelled during delay", zap.Int64("task_id", task.ID))
			return ctx.Err()
		default:
		}

		// Rate limiting - use rateLimiter instead of fixed sleep
		// The rateLimiter already handles timing, but we can add a small delay for safety
		if err := eac.rateLimiter.WaitForDomain(ctx, "www.embroiderydesigns.com"); err != nil {
			eac.logger.Warn("Rate limiter error", zap.Error(err))
			time.Sleep(500 * time.Millisecond)
		}

		// Add random delay between requests (1-3 seconds) to avoid detection
		randomDelay := time.Duration(1000+rand.Intn(2000)) * time.Millisecond
		select {
		case <-ctx.Done():
			eac.logger.Info("Crawl cancelled during random delay", zap.Int64("task_id", task.ID))
			return ctx.Err()
		case <-time.After(randomDelay):
		}
	}

	// Final stats report
	elapsed := time.Since(stats.StartTime)
	completionMsg := fmt.Sprintf("Crawl completed: %d products processed (Success: %d, Errors: %d) in %v",
		totalProcessed, stats.SuccessCount, stats.ErrorCount, elapsed.Round(time.Second))

	eac.logger.Info("Crawl completed",
		zap.Int64("task_id", task.ID),
		zap.Int("total_processed", totalProcessed),
		zap.Int64("success_count", stats.SuccessCount),
		zap.Int64("error_count", stats.ErrorCount),
		zap.Duration("elapsed", elapsed),
	)
	
	if eac.logCallback != nil {
		eac.logCallback(task.ID, "info", completionMsg)
	}

	// Clear last_from from config since crawl is complete
	if taskConfig != nil {
		delete(taskConfig, "last_from")
		if configBytes, err := json.Marshal(taskConfig); err == nil {
			_ = eac.repository.UpdateTaskConfig(ctx, task.ID, string(configBytes))
		}
	}

	// Start periodic monitoring
	go eac.StartPeriodicMonitoring(ctx, task, totalAvailable)

	return nil
}

func (eac *EmbroideryAPICrawler) saveProductsBatch(
	ctx context.Context,
	batch []map[string]interface{},
	batchIDs []string,
	stats *CrawlStats,
) (int64, int64) {
	var successCount int64
	var errorCount int64

	for i, source := range batch {
		// Check context before processing each product
		select {
		case <-ctx.Done():
			eac.logger.Info("Crawl cancelled during batch save", zap.Int64("batch_size", int64(len(batch))))
			return successCount, errorCount
		default:
		}

		elasticID := batchIDs[i]
		
		// Retry saving individual products
		err := utils.Retry(ctx, eac.retryConfig, func() error {
			return eac.saveProduct(ctx, elasticID, source)
		})

		if err != nil {
			errorCount++
			errMsg := fmt.Sprintf("Failed to save product %s: %v", elasticID, err)
			eac.logger.Error("Failed to save product",
				zap.String("elastic_id", elasticID),
				zap.Error(err),
			)
			if len(stats.Errors) < 100 { // Limit error log size
				stats.Errors = append(stats.Errors, errMsg)
			}
		} else {
			successCount++
		}
	}

	return successCount, errorCount
}

func (eac *EmbroideryAPICrawler) StartPeriodicMonitoring(ctx context.Context, task *storage.Task, lastTotal int64) {
	checkInterval := eac.config.Embroidery.CheckInterval
	if checkInterval == 0 {
		checkInterval = 6 * time.Hour
	}

	eac.mu.Lock()
	if _, exists := eac.periodicMonitors[task.ID]; exists {
		eac.mu.Unlock()
		return // Already monitoring
	}
	monitorCtx, cancel := context.WithCancel(ctx)
	eac.periodicMonitors[task.ID] = cancel
	eac.mu.Unlock()

	logMsg := fmt.Sprintf("Periodic monitoring started: checking every %v", checkInterval)
	eac.logger.Info(logMsg, zap.Int64("task_id", task.ID))
	if eac.logCallback != nil {
		eac.logCallback(task.ID, "info", logMsg)
	}

	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()
		defer func() {
			eac.mu.Lock()
			delete(eac.periodicMonitors, task.ID)
			eac.mu.Unlock()
		}()

		for {
			select {
			case <-monitorCtx.Done():
				eac.logger.Info("Periodic monitoring stopped", zap.Int64("task_id", task.ID))
				return
			case <-ticker.C:
				// Check if there are new products
				eac.checkForNewProducts(monitorCtx, task, lastTotal)
			}
		}
	}()
}

func (eac *EmbroideryAPICrawler) checkForNewProducts(ctx context.Context, task *storage.Task, lastTotal int64) {
	logMsg := "Checking for new products..."
	eac.logger.Info(logMsg, zap.Int64("task_id", task.ID))
	if eac.logCallback != nil {
		eac.logCallback(task.ID, "info", logMsg)
	}

	// Get current total from API
	pageSize := eac.config.Embroidery.PageSize
	if pageSize == 0 {
		pageSize = 120
	}

	payload := eac.createPayload(0, 1, eac.loadPayloadOverrides(ctx)) // Just get total count
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		eac.logger.Error("Failed to create check payload", zap.Error(err))
		return
	}

	resp, err := eac.makeRequest(ctx, payloadBytes)
	if err != nil {
		eac.logger.Error("Failed to check for new products", zap.Error(err))
		return
	}

	var esResp ElasticsearchResponse
	if err := json.Unmarshal(resp, &esResp); err != nil {
		eac.logger.Error("Failed to parse check response", zap.Error(err))
		return
	}

	currentTotal := esResp.Hits.Total.Value
	if currentTotal > lastTotal {
		newCount := currentTotal - lastTotal
		logMsg := fmt.Sprintf("Found %d new products. Starting incremental crawl...", newCount)
		eac.logger.Info(logMsg,
			zap.Int64("task_id", task.ID),
			zap.Int64("last_total", lastTotal),
			zap.Int64("current_total", currentTotal),
			zap.Int64("new_count", newCount),
		)
		if eac.logCallback != nil {
			eac.logCallback(task.ID, "info", logMsg)
		}

		// Start incremental crawl
		go eac.IncrementalCrawl(ctx, task, int(lastTotal))
	} else {
		logMsg := fmt.Sprintf("No new products found. Current total: %d", currentTotal)
		eac.logger.Info(logMsg, zap.Int64("task_id", task.ID))
		if eac.logCallback != nil {
			eac.logCallback(task.ID, "info", logMsg)
		}
	}
}

func (eac *EmbroideryAPICrawler) IncrementalCrawl(ctx context.Context, task *storage.Task, from int) {
	logMsg := fmt.Sprintf("Starting incremental crawl from position %d", from)
	eac.logger.Info(logMsg, zap.Int64("task_id", task.ID))
	if eac.logCallback != nil {
		eac.logCallback(task.ID, "info", logMsg)
	}

	pageSize := eac.config.Embroidery.PageSize
	if pageSize == 0 {
		pageSize = 120
	}

	stats := &CrawlStats{
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Errors:     make([]string, 0),
	}

	payloadOverrides := eac.loadPayloadOverrides(ctx)

	totalProcessed := 0
	const batchSize = 50

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		payload := eac.createPayload(from, pageSize, payloadOverrides)
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			eac.logger.Error("Failed to marshal payload", zap.Error(err))
			return
		}

		resp, err := eac.makeRequest(ctx, payloadBytes)
		if err != nil {
			eac.logger.Error("Failed to make request", zap.Error(err))
			return
		}

		var esResp ElasticsearchResponse
		if err := json.Unmarshal(resp, &esResp); err != nil {
			eac.logger.Error("Failed to parse response", zap.Error(err))
			return
		}

		productsCount := len(esResp.Hits.Hits)
		if productsCount == 0 {
			break
		}

		// Save products in batches
		batch := make([]map[string]interface{}, 0, batchSize)
		batchIDs := make([]string, 0, batchSize)

		for i, hit := range esResp.Hits.Hits {
			batch = append(batch, hit.Source)
			batchIDs = append(batchIDs, hit.ID)

			if len(batch) >= batchSize || i == len(esResp.Hits.Hits)-1 {
				successCount, errCount := eac.saveProductsBatch(ctx, batch, batchIDs, stats)
				stats.SuccessCount += successCount
				stats.ErrorCount += errCount
				
				batch = make([]map[string]interface{}, 0, batchSize)
				batchIDs = make([]string, 0, batchSize)
			}
		}

		totalProcessed += productsCount
		logMsg := fmt.Sprintf("Incremental crawl: %d products processed (total: %d)", productsCount, totalProcessed)
		eac.logger.Info(logMsg, zap.Int64("task_id", task.ID))
		if eac.logCallback != nil {
			eac.logCallback(task.ID, "info", logMsg)
		}

		if productsCount < pageSize {
			break
		}

		from += pageSize

		if err := eac.rateLimiter.WaitForDomain(ctx, "www.embroiderydesigns.com"); err != nil {
			time.Sleep(500 * time.Millisecond)
		}
	}

	logMsg = fmt.Sprintf("Incremental crawl completed: %d products processed", totalProcessed)
	eac.logger.Info(logMsg, zap.Int64("task_id", task.ID))
	if eac.logCallback != nil {
		eac.logCallback(task.ID, "info", logMsg)
	}
}

func (eac *EmbroideryAPICrawler) createPayload(from, size int, overrides map[string]interface{}) map[string]interface{} {
	return BuildEmbroideryPayload(from, size, overrides)
}

func (eac *EmbroideryAPICrawler) makeRequest(ctx context.Context, payload []byte) ([]byte, error) {
	// Get proxy if enabled
	var currentProxy *storage.Proxy
	if eac.config.Proxy.Enabled {
		var err error
		currentProxy, err = eac.proxyManager.GetProxy()
		if err != nil {
			eac.logger.Warn("Failed to get proxy, continuing without proxy", zap.Error(err))
		}
	}

	// Get HTTP client
	client, err := eac.proxyManager.GetHTTPClient(ctx, currentProxy)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", eac.baseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Generate fingerprint profile
	profile := fingerprint.GenerateProfile()
	
	// Apply fingerprint headers
	fingerprint.ApplyHeaders(req, profile)
	
	// Merge with API-specific headers (API headers take precedence)
	for k, v := range eac.apiHeaders {
		req.Header.Set(k, v)
	}
	// Always request compressed payloads (API uses gzip by default)
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	}

	// Set cookies if provided
	if eac.cookies != "" {
		req.Header.Set("Cookie", eac.cookies)
	}

	// Rate limiting
	if err := eac.rateLimiter.WaitForDomain(ctx, req.URL.Host); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Execute request with retry
	var resp *http.Response
	err = utils.Retry(ctx, eac.retryConfig, func() error {
		var retryErr error
		resp, retryErr = client.Do(req)
		if retryErr != nil {
			if currentProxy != nil {
				eac.proxyManager.ReportProxyFailure(ctx, currentProxy)
				currentProxy, _ = eac.proxyManager.GetProxy()
				if currentProxy != nil {
					client, _ = eac.proxyManager.GetHTTPClient(ctx, currentProxy)
				}
			}
			return retryErr
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to execute request after retries: %w", err)
	}
	defer resp.Body.Close()

	var reader io.Reader = resp.Body
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		if gzipReader, gzErr := gzip.NewReader(resp.Body); gzErr == nil {
			defer gzipReader.Close()
			reader = gzipReader
		} else {
			return nil, fmt.Errorf("failed to init gzip reader: %w", gzErr)
		}
	case "br":
		reader = brotli.NewReader(resp.Body)
	case "deflate":
		if defReader, defErr := zlib.NewReader(resp.Body); defErr == nil {
			defer defReader.Close()
			reader = defReader
		} else {
			return nil, fmt.Errorf("failed to init deflate reader: %w", defErr)
		}
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(reader)
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read response body
	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return bodyBytes, nil
}

func (eac *EmbroideryAPICrawler) saveProduct(ctx context.Context, elasticID string, source map[string]interface{}) error {
	// Convert source to JSON for raw_data
	rawDataBytes, _ := json.Marshal(source)
	rawData := string(rawDataBytes)

	// Extract fields
	product := &storage.Product{
		ElasticID: elasticID,
		RawData:   &rawData,
		Status:    storage.ProductStatusPending,
	}

	// Helper function to extract string value
	getString := func(key string) *string {
		if val, ok := source[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				return &str
			}
		}
		return nil
	}

	// Helper function to extract float value
	getFloat := func(key string) *float64 {
		if val, ok := source[key]; ok {
			switch v := val.(type) {
			case float64:
				return &v
			case int:
				f := float64(v)
				return &f
			case int64:
				f := float64(v)
				return &f
			}
		}
		return nil
	}

	// Helper function to extract int value
	getInt := func(key string) *int {
		if val, ok := source[key]; ok {
			switch v := val.(type) {
			case int:
				return &v
			case int64:
				i := int(v)
				return &i
			case float64:
				i := int(v)
				return &i
			}
		}
		return nil
	}

	// Helper function to extract bool value
	getBool := func(key string) bool {
		if val, ok := source[key]; ok {
			if b, ok := val.(bool); ok {
				return b
			}
		}
		return false
	}

	// Helper function to extract time value
	getTime := func(key string) *time.Time {
		if val, ok := source[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				// Try different time formats
				formats := []string{
					time.RFC3339,
					"2006-01-02T15:04:05",
					"2006-01-02T15:04:05.0000000",
					"2006-01-02",
				}
				for _, format := range formats {
					if t, err := time.Parse(format, str); err == nil {
						return &t
					}
				}
			}
		}
		return nil
	}

	// Helper function to extract JSON array as string
	getJSONArray := func(key string) *string {
		if val, ok := source[key]; ok {
			if bytes, err := json.Marshal(val); err == nil {
				str := string(bytes)
				return &str
			}
		}
		return nil
	}

	// Extract all fields
	product.ProductID = getString("productId")
	product.ItemID = getString("itemId")
	product.Name = getString("name")
	product.Brand = getString("brand")
	product.Catalog = getString("catalog")
	product.Artist = getString("artist")
	product.Rating = getFloat("rating")
	product.ListPrice = getFloat("listPrice")
	product.SalePrice = getFloat("salePrice")
	product.ClubPrice = getFloat("clubPrice")
	product.SaleRank = getInt("saleRank")
	product.CustomerInterestIndex = getInt("customerInterestIndex")
	product.InStock = getBool("inStock")
	product.IsActive = getBool("isActive")
	product.IsBuyable = getBool("isBuyable")
	product.Licensed = getBool("licensed")
	product.IsApplique = getBool("isApplique")
	product.IsCrossStitch = getBool("isCrossStitch")
	product.IsPDFAvailable = getBool("isPDFAvailable")
	product.IsFSL = getBool("isFSL")
	product.IsHeatTransfer = getBool("isHeatTransfer")
	product.IsDesignUsedInProject = getBool("isDesignUsedInProject")
	product.InCustomPack = getBool("inCustomPack")
	product.DefinitionName = getString("definitionName")
	product.ProductType = getString("productType")
	product.GTIN = getString("gtin")
	product.ColorSequence = getString("colorSequence")
	product.DesignKeywords = getString("designKeywords")
	product.Categories = getString("categories")
	product.CategoriesList = getJSONArray("categoriesList")
	product.Keywords = getJSONArray("keywords")
	product.Sales = getString("sales")
	product.SalesList = getJSONArray("salesList")
	product.SaleEndDate = getTime("saleEndDate")
	product.YearCreated = getTime("yearCreated")
	product.AppliedDiscountID = getInt("appliedDiscountId")
	product.IsMultipleVariantsAvailable = getBool("isMultipleVariantsAvailable")
	product.Variants = getJSONArray("variants")

	// Save to database
	return eac.repository.UpsertProduct(ctx, product)
}

func (eac *EmbroideryAPICrawler) loadPayloadOverrides(ctx context.Context) map[string]interface{} {
	overrides, _, err := eac.repository.GetEmbroideryPayloadOverrides(ctx)
	if err != nil {
		eac.logger.Warn("Failed to load embroidery payload overrides, falling back to defaults", zap.Error(err))
		return nil
	}
	if overrides == nil {
		return map[string]interface{}{}
	}
	return overrides
}

func (eac *EmbroideryAPICrawler) StopPeriodicMonitoring(taskID int64) {
	eac.mu.Lock()
	defer eac.mu.Unlock()
	
	if cancel, exists := eac.periodicMonitors[taskID]; exists {
		cancel()
		delete(eac.periodicMonitors, taskID)
		eac.logger.Info("Periodic monitoring stopped", zap.Int64("task_id", taskID))
	}
}
