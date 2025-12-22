package crawler

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"embroidery-designs/internal/storage"
	"embroidery-designs/internal/utils"
)

type WorkerPool struct {
	workers    int
	jobChan    chan *storage.Task
	resultChan chan *CrawlResult
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *zap.Logger
	crawler    Crawler
}

func NewWorkerPool(workers int, crawler Crawler) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		workers:    workers,
		jobChan:    make(chan *storage.Task, workers*2),
		resultChan: make(chan *CrawlResult, workers*2),
		ctx:        ctx,
		cancel:     cancel,
		logger:     utils.GetLogger(),
		crawler:    crawler,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	
	wp.logger.Info("Worker pool started", zap.Int("workers", wp.workers))
}

func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.jobChan)
	wp.wg.Wait()
	close(wp.resultChan)
	wp.logger.Info("Worker pool stopped")
}

func (wp *WorkerPool) Submit(task *storage.Task) {
	select {
	case wp.jobChan <- task:
	case <-wp.ctx.Done():
		return
	}
}

func (wp *WorkerPool) GetResult() <-chan *CrawlResult {
	return wp.resultChan
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	wp.logger.Debug("Worker started", zap.Int("worker_id", id))
	
	for {
		select {
		case task, ok := <-wp.jobChan:
			if !ok {
				wp.logger.Debug("Worker stopping", zap.Int("worker_id", id))
				return
			}
			
			wp.logger.Debug("Worker processing task",
				zap.Int("worker_id", id),
				zap.Int64("task_id", task.ID),
			)
			
			if err := wp.crawler.Crawl(wp.ctx, task); err != nil {
				wp.logger.Error("Worker failed to crawl",
					zap.Int("worker_id", id),
					zap.Int64("task_id", task.ID),
					zap.Error(err),
				)
			}
			
		case <-wp.ctx.Done():
			wp.logger.Debug("Worker context cancelled", zap.Int("worker_id", id))
			return
		}
	}
}

