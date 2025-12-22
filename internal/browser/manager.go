package browser

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/fingerprint"
	"embroidery-designs/internal/utils"
)

type Manager struct {
	launcher *Launcher
	config   *config.Config
	logger   *zap.Logger
	mu       sync.RWMutex
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		launcher: NewLauncher(cfg),
		config:   cfg,
		logger:   utils.GetLogger(),
	}
}

func (m *Manager) GetLauncher() *Launcher {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.launcher
}

func (m *Manager) CreateContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	return m.launcher.CreateContext(ctx)
}

func (m *Manager) GenerateProfile() *fingerprint.BrowserProfile {
	return fingerprint.GenerateProfile()
}

func (m *Manager) SetProfile(profile *fingerprint.BrowserProfile) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.launcher.SetProfile(profile)
}

