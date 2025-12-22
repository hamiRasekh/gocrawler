package browser

import (
	"context"
	"fmt"
	"os"

	"github.com/chromedp/chromedp"
	"go.uber.org/zap"
	"embroidery-designs/internal/config"
	"embroidery-designs/internal/fingerprint"
	"embroidery-designs/internal/utils"
)

type Launcher struct {
	config  *config.Config
	logger  *zap.Logger
	profile *fingerprint.BrowserProfile
}

func NewLauncher(cfg *config.Config) *Launcher {
	return &Launcher{
		config:  cfg,
		logger:  utils.GetLogger(),
		profile: fingerprint.GenerateProfile(),
	}
}

func (l *Launcher) CreateContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"), // Use new headless mode
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-features", "VizDisplayCompositor"),
		chromedp.UserAgent(l.profile.UserAgent),
		chromedp.WindowSize(l.profile.ViewportWidth, l.profile.ViewportHeight),
	)
	
	// Set Chromium path for Alpine/Docker (if environment variable is set)
	if chromiumPath := os.Getenv("CHROMIUM_PATH"); chromiumPath != "" {
		opts = append(opts, chromedp.ExecPath(chromiumPath))
	}
	
	if l.config.Browser.UserDataDir != "" {
		opts = append(opts, chromedp.UserDataDir(l.config.Browser.UserDataDir))
	}
	
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	
	ctx, cancel2 := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(format string, v ...interface{}) {
		l.logger.Debug(fmt.Sprintf(format, v...))
	}))
	
	return ctx, func() {
		cancel2()
		cancel()
	}, nil
}

func (l *Launcher) Navigate(ctx context.Context, url string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, l.config.Browser.Timeout)
	defer cancel()
	
	var html string
	err := chromedp.Run(timeoutCtx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &html),
	)
	
	if err != nil {
		return fmt.Errorf("failed to navigate: %w", err)
	}
	
	return nil
}

func (l *Launcher) GetPageContent(ctx context.Context) (string, error) {
	var html string
	err := chromedp.Run(ctx,
		chromedp.OuterHTML("html", &html),
	)
	
	if err != nil {
		return "", fmt.Errorf("failed to get page content: %w", err)
	}
	
	return html, nil
}

func (l *Launcher) ExecuteScript(ctx context.Context, script string) (interface{}, error) {
	var result interface{}
	err := chromedp.Run(ctx,
		chromedp.Evaluate(script, &result),
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to execute script: %w", err)
	}
	
	return result, nil
}

func (l *Launcher) SetProfile(profile *fingerprint.BrowserProfile) {
	l.profile = profile
}

func (l *Launcher) GetProfile() *fingerprint.BrowserProfile {
	return l.profile
}

