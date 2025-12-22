package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/net/proxy"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
)

type Manager struct {
	pool          *ProxyPool
	healthChecker *HealthChecker
	config        *config.Config
	logger        *zap.Logger
	mu            sync.RWMutex
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

func NewManager(repo *storage.Repository, cfg *config.Config) *Manager {
	pool := NewProxyPool(repo)
	healthChecker := NewHealthChecker(repo, cfg)
	
	return &Manager{
		pool:          pool,
		healthChecker: healthChecker,
		config:        cfg,
		logger:        utils.GetLogger(),
		stopChan:      make(chan struct{}),
	}
}

func (m *Manager) Start(ctx context.Context) error {
	// Load initial proxies
	if err := m.pool.LoadProxies(ctx); err != nil {
		return fmt.Errorf("failed to load proxies: %w", err)
	}
	
	// Start health check routine
	if m.config.Proxy.Enabled {
		m.wg.Add(1)
		go m.healthCheckRoutine(ctx)
	}
	
	m.logger.Info("Proxy manager started")
	return nil
}

func (m *Manager) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	m.logger.Info("Proxy manager stopped")
}

func (m *Manager) GetProxy() (*storage.Proxy, error) {
	if !m.config.Proxy.Enabled {
		return nil, nil // No proxy needed
	}
	
	return m.pool.GetProxy()
}

func (m *Manager) GetHTTPClient(ctx context.Context, proxy *storage.Proxy) (*http.Client, error) {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	
	if proxy != nil {
		if err := m.setProxy(transport, proxy); err != nil {
			return nil, fmt.Errorf("failed to set proxy: %w", err)
		}
	}
	
	client := &http.Client{
		Transport: transport,
		Timeout:   m.config.Crawler.RequestTimeout,
	}
	
	return client, nil
}

func (m *Manager) setProxy(transport *http.Transport, p *storage.Proxy) error {
	switch p.Type {
	case "http", "https":
		var proxyURL string
		if p.Username != nil && p.Password != nil {
			proxyURL = fmt.Sprintf("%s://%s:%s@%s:%d",
				p.Type, *p.Username, *p.Password, p.Host, p.Port)
		} else {
			proxyURL = fmt.Sprintf("%s://%s:%d", p.Type, p.Host, p.Port)
		}
		
		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			return fmt.Errorf("failed to parse proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsedURL)
		
	case "socks5":
		var auth *proxy.Auth
		if p.Username != nil && p.Password != nil {
			auth = &proxy.Auth{
				User:     *p.Username,
				Password: *p.Password,
			}
		}
		
		dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("%s:%d", p.Host, p.Port), auth, proxy.Direct)
		if err != nil {
			return fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
		}
		
		transport.DialContext = dialer.(proxy.ContextDialer).DialContext
		
	default:
		return fmt.Errorf("unsupported proxy type: %s", p.Type)
	}
	
	return nil
}

func (m *Manager) ReportProxyFailure(ctx context.Context, proxy *storage.Proxy) {
	if proxy == nil {
		return
	}
	
	if err := m.pool.repository.UpdateProxyHealth(ctx, proxy.ID, false); err != nil {
		m.logger.Error("Failed to report proxy failure",
			zap.Int64("proxy_id", proxy.ID),
			zap.Error(err),
		)
	}
	
	// Reload proxies to get updated health status
	_ = m.pool.LoadProxies(ctx)
}

func (m *Manager) ReloadProxies(ctx context.Context) error {
	return m.pool.LoadProxies(ctx)
}

func (m *Manager) healthCheckRoutine(ctx context.Context) {
	defer m.wg.Done()
	
	ticker := time.NewTicker(m.config.Proxy.HealthCheckInterval)
	defer ticker.Stop()
	
	// Initial health check
	m.performHealthCheck(ctx)
	
	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			m.performHealthCheck(ctx)
		}
	}
}

func (m *Manager) performHealthCheck(ctx context.Context) {
	proxies := m.pool.GetAllProxies()
	if len(proxies) == 0 {
		return
	}
	
	m.logger.Info("Starting proxy health check", zap.Int("count", len(proxies)))
	m.healthChecker.CheckAllProxies(ctx, proxies)
	
	// Reload proxies after health check
	if err := m.pool.LoadProxies(ctx); err != nil {
		m.logger.Error("Failed to reload proxies after health check", zap.Error(err))
	}
}

func (m *Manager) GetPool() *ProxyPool {
	return m.pool
}

func (m *Manager) TestProxy(ctx context.Context, proxy *storage.Proxy) (bool, error) {
	return m.healthChecker.CheckProxy(ctx, proxy)
}

