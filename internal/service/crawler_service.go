package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"embroidery-designs/internal/crawler"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
	"embroidery-designs/internal/websocket"
	"go.uber.org/zap"
)

type CrawlerService struct {
	repository        *storage.Repository
	apiCrawler        *crawler.APICrawler
	browserCrawler    *crawler.BrowserCrawler
	embroideryCrawler *crawler.EmbroideryAPICrawler
	webCrawler        *crawler.WebCrawler
	workerPool        *crawler.WorkerPool
	activeTasks       map[int64]context.CancelFunc
	wsHub             *websocket.Hub
	mu                sync.RWMutex
	logger            *zap.Logger
}

func NewCrawlerService(
	repo *storage.Repository,
	apiCrawler *crawler.APICrawler,
	browserCrawler *crawler.BrowserCrawler,
	embroideryCrawler *crawler.EmbroideryAPICrawler,
	wsHub *websocket.Hub,
) *CrawlerService {
	webCrawler := crawler.NewWebCrawler(apiCrawler, browserCrawler)
	workerPool := crawler.NewWorkerPool(10, webCrawler)

	service := &CrawlerService{
		repository:        repo,
		apiCrawler:        apiCrawler,
		browserCrawler:    browserCrawler,
		embroideryCrawler: embroideryCrawler,
		webCrawler:        webCrawler,
		workerPool:        workerPool,
		activeTasks:       make(map[int64]context.CancelFunc),
		wsHub:             wsHub,
		logger:            utils.GetLogger(),
	}

	if service.embroideryCrawler != nil {
		service.embroideryCrawler.SetLogCallback(func(taskID int64, level, message string) {
			service.BroadcastLog(taskID, level, message)
		})
	}

	return service
}

func (cs *CrawlerService) BroadcastLog(taskID int64, level, message string) {
	if cs.wsHub != nil {
		cs.wsHub.Broadcast(websocket.Message{
			Type:    "log",
			TaskID:  &taskID,
			Level:   level,
			Message: message,
		})
	}
}

func (cs *CrawlerService) BroadcastTaskStatus(taskID int64, status string) {
	if cs.wsHub != nil {
		cs.wsHub.Broadcast(websocket.Message{
			Type:   "task_status",
			TaskID: &taskID,
			Data: map[string]interface{}{
				"status": status,
			},
		})
	}
}

func (cs *CrawlerService) StartTask(ctx context.Context, taskID int64) error {
	task, err := cs.repository.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if task.Status == storage.TaskStatusRunning {
		return fmt.Errorf("task is already running")
	}

	// Update status
	if err := cs.repository.UpdateTaskStatus(ctx, taskID, storage.TaskStatusRunning); err != nil {
		return err
	}

	// Broadcast task status
	cs.BroadcastTaskStatus(taskID, string(storage.TaskStatusRunning))
	cs.BroadcastLog(taskID, "info", fmt.Sprintf("Task %d started", taskID))

	// Create context for this task that lives beyond the HTTP request scope
	taskCtx, cancel := context.WithCancel(context.Background())
	cs.mu.Lock()
	cs.activeTasks[taskID] = cancel
	cs.mu.Unlock()

	// Start worker pool if not started
	cs.workerPool.Start()

	// Submit task to worker pool
	go func() {
		defer func() {
			cs.mu.Lock()
			delete(cs.activeTasks, taskID)
			cs.mu.Unlock()
		}()

		var err error

		// Check if this is an embroidery API task
		var taskConfig map[string]interface{}
		if task.Config != "" {
			if parseErr := json.Unmarshal([]byte(task.Config), &taskConfig); parseErr == nil {
				if taskType, ok := taskConfig["crawler_type"].(string); ok && taskType == "embroidery_api" {
					// Use embroidery API crawler
					if cs.embroideryCrawler != nil {
						err = cs.embroideryCrawler.CrawlAll(taskCtx, task)
					} else {
						err = fmt.Errorf("embroidery crawler not initialized")
					}
				} else {
					// Use regular web crawler
					err = cs.webCrawler.Crawl(taskCtx, task)
				}
			} else {
				// Use regular web crawler
				err = cs.webCrawler.Crawl(taskCtx, task)
			}
		} else {
			// Use regular web crawler
			err = cs.webCrawler.Crawl(taskCtx, task)
		}

		if err != nil {
			cs.logger.Error("Task failed",
				zap.Int64("task_id", taskID),
				zap.Error(err),
			)
			cs.BroadcastLog(taskID, "error", fmt.Sprintf("Task failed: %v", err))
			_ = cs.repository.UpdateTaskStatus(taskCtx, taskID, storage.TaskStatusFailed)
			cs.BroadcastTaskStatus(taskID, string(storage.TaskStatusFailed))
			return
		}

		cs.BroadcastLog(taskID, "info", fmt.Sprintf("Task %d completed successfully", taskID))
		_ = cs.repository.UpdateTaskStatus(taskCtx, taskID, storage.TaskStatusCompleted)
		cs.BroadcastTaskStatus(taskID, string(storage.TaskStatusCompleted))
	}()

	return nil
}

func (cs *CrawlerService) StopTask(ctx context.Context, taskID int64) error {
	cs.mu.Lock()
	cancel, exists := cs.activeTasks[taskID]
	cs.mu.Unlock()

	if exists {
		cancel()
	}

	// Stop periodic monitoring for embroidery crawler
	if cs.embroideryCrawler != nil {
		cs.embroideryCrawler.StopPeriodicMonitoring(taskID)
	}

	cs.BroadcastLog(taskID, "info", fmt.Sprintf("Task %d stopped", taskID))
	cs.BroadcastTaskStatus(taskID, string(storage.TaskStatusStopped))
	return cs.repository.UpdateTaskStatus(ctx, taskID, storage.TaskStatusStopped)
}

func (cs *CrawlerService) PauseTask(ctx context.Context, taskID int64) error {
	return cs.repository.UpdateTaskStatus(ctx, taskID, storage.TaskStatusPaused)
}

func (cs *CrawlerService) ResumeTask(ctx context.Context, taskID int64) error {
	return cs.StartTask(ctx, taskID)
}
