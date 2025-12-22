package proxy

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
)

type ProxyPool struct {
	proxies    []*storage.Proxy
	mu         sync.RWMutex
	repository *storage.Repository
	logger     *zap.Logger
	rand       *rand.Rand
}

func NewProxyPool(repo *storage.Repository) *ProxyPool {
	return &ProxyPool{
		proxies:    make([]*storage.Proxy, 0),
		repository: repo,
		logger:     utils.GetLogger(),
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (p *ProxyPool) LoadProxies(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	proxies, err := p.repository.GetActiveProxies(ctx)
	if err != nil {
		return fmt.Errorf("failed to load proxies: %w", err)
	}
	
	p.proxies = proxies
	p.logger.Info("Loaded proxies into pool", zap.Int("count", len(p.proxies)))
	
	return nil
}

func (p *ProxyPool) GetProxy() (*storage.Proxy, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if len(p.proxies) == 0 {
		return nil, fmt.Errorf("no proxies available")
	}
	
	// Get random proxy from pool
	idx := p.rand.Intn(len(p.proxies))
	return p.proxies[idx], nil
}

func (p *ProxyPool) GetProxyByIndex(idx int) (*storage.Proxy, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if idx < 0 || idx >= len(p.proxies) {
		return nil, fmt.Errorf("proxy index out of range")
	}
	
	return p.proxies[idx], nil
}

func (p *ProxyPool) GetAllProxies() []*storage.Proxy {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	proxies := make([]*storage.Proxy, len(p.proxies))
	copy(proxies, p.proxies)
	return proxies
}

func (p *ProxyPool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.proxies)
}

func (p *ProxyPool) AddProxy(proxy *storage.Proxy) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.proxies = append(p.proxies, proxy)
}

func (p *ProxyPool) RemoveProxy(id int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for i, proxy := range p.proxies {
		if proxy.ID == id {
			p.proxies = append(p.proxies[:i], p.proxies[i+1:]...)
			break
		}
	}
}

