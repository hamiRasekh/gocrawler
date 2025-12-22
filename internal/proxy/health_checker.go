package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
)

type HealthChecker struct {
	repository *storage.Repository
	config     *config.Config
	logger     *zap.Logger
	client     *http.Client
}

func NewHealthChecker(repo *storage.Repository, cfg *config.Config) *HealthChecker {
	return &HealthChecker{
		repository: repo,
		config:     cfg,
		logger:     utils.GetLogger(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (hc *HealthChecker) CheckProxy(ctx context.Context, proxy *storage.Proxy) (bool, error) {
	proxyURL, err := hc.buildProxyURL(proxy)
	if err != nil {
		return false, err
	}
	
	transport := &http.Transport{}
	if err := hc.setProxy(transport, proxyURL); err != nil {
		return false, err
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
	
	// Test with a simple request
	testURL := "https://httpbin.org/ip"
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return false, err
	}
	
	resp, err := client.Do(req)
	if err != nil {
		hc.logger.Debug("Proxy health check failed",
			zap.Int64("proxy_id", proxy.ID),
			zap.String("proxy", proxyURL),
			zap.Error(err),
		)
		return false, nil
	}
	defer resp.Body.Close()
	
	isHealthy := resp.StatusCode == http.StatusOK
	hc.logger.Debug("Proxy health check completed",
		zap.Int64("proxy_id", proxy.ID),
		zap.String("proxy", proxyURL),
		zap.Bool("healthy", isHealthy),
		zap.Int("status_code", resp.StatusCode),
	)
	
	return isHealthy, nil
}

func (hc *HealthChecker) CheckAllProxies(ctx context.Context, proxies []*storage.Proxy) {
	for _, proxy := range proxies {
		healthy, err := hc.CheckProxy(ctx, proxy)
		if err != nil {
			hc.logger.Warn("Error checking proxy health",
				zap.Int64("proxy_id", proxy.ID),
				zap.Error(err),
			)
			continue
		}
		
		if err := hc.repository.UpdateProxyHealth(ctx, proxy.ID, healthy); err != nil {
			hc.logger.Error("Failed to update proxy health",
				zap.Int64("proxy_id", proxy.ID),
				zap.Error(err),
			)
		}
		
		// Mark as inactive if failures exceed threshold
		if !healthy && proxy.FailureCount+1 >= hc.config.Proxy.MaxFailures {
			hc.logger.Warn("Proxy exceeded failure threshold, marking as inactive",
				zap.Int64("proxy_id", proxy.ID),
				zap.Int("failure_count", proxy.FailureCount+1),
			)
		}
	}
}

func (hc *HealthChecker) buildProxyURL(proxy *storage.Proxy) (string, error) {
	var proxyURL string
	
	switch proxy.Type {
	case "http", "https":
		if proxy.Username != nil && proxy.Password != nil {
			proxyURL = fmt.Sprintf("%s://%s:%s@%s:%d",
				proxy.Type, *proxy.Username, *proxy.Password, proxy.Host, proxy.Port)
		} else {
			proxyURL = fmt.Sprintf("%s://%s:%d", proxy.Type, proxy.Host, proxy.Port)
		}
	case "socks5":
		if proxy.Username != nil && proxy.Password != nil {
			proxyURL = fmt.Sprintf("socks5://%s:%s@%s:%d",
				*proxy.Username, *proxy.Password, proxy.Host, proxy.Port)
		} else {
			proxyURL = fmt.Sprintf("socks5://%s:%d", proxy.Host, proxy.Port)
		}
	default:
		return "", fmt.Errorf("unsupported proxy type: %s", proxy.Type)
	}
	
	return proxyURL, nil
}

func (hc *HealthChecker) setProxy(transport *http.Transport, proxyURL string) error {
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("failed to parse proxy URL: %w", err)
	}
	
	transport.Proxy = http.ProxyURL(parsedURL)
	return nil
}

