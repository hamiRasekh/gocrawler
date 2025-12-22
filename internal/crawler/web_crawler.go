package crawler

import (
	"context"
	"embroidery-designs/internal/storage"
)

// WebCrawler can use either API crawler or Browser crawler based on task type
type WebCrawler struct {
	apiCrawler     *APICrawler
	browserCrawler *BrowserCrawler
}

func NewWebCrawler(apiCrawler *APICrawler, browserCrawler *BrowserCrawler) *WebCrawler {
	return &WebCrawler{
		apiCrawler:     apiCrawler,
		browserCrawler: browserCrawler,
	}
}

func (wc *WebCrawler) Crawl(ctx context.Context, task *storage.Task) error {
	if task.Type == "api" {
		return wc.apiCrawler.Crawl(ctx, task)
	}
	return wc.browserCrawler.Crawl(ctx, task)
}

func (wc *WebCrawler) Stop(ctx context.Context, taskID int64) error {
	// Try both, one will work
	_ = wc.apiCrawler.Stop(ctx, taskID)
	_ = wc.browserCrawler.Stop(ctx, taskID)
	return nil
}

func (wc *WebCrawler) Pause(ctx context.Context, taskID int64) error {
	_ = wc.apiCrawler.Pause(ctx, taskID)
	_ = wc.browserCrawler.Pause(ctx, taskID)
	return nil
}

func (wc *WebCrawler) Resume(ctx context.Context, taskID int64) error {
	_ = wc.apiCrawler.Resume(ctx, taskID)
	_ = wc.browserCrawler.Resume(ctx, taskID)
	return nil
}

