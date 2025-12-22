package crawler

import (
	"context"
	"embroidery-designs/internal/storage"
)

type Crawler interface {
	Crawl(ctx context.Context, task *storage.Task) error
	Stop(ctx context.Context, taskID int64) error
	Pause(ctx context.Context, taskID int64) error
	Resume(ctx context.Context, taskID int64) error
}

type CrawlResult struct {
	URL         string
	Method      string
	StatusCode  int
	Headers     map[string]string
	Body        string
	ResponseTime int
	ProxyUsed   *string
}

