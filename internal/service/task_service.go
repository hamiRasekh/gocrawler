package service

import (
	"context"
	"encoding/json"
	"fmt"

	"embroidery-designs/internal/storage"
)

type TaskService struct {
	repository *storage.Repository
}

func NewTaskService(repo *storage.Repository) *TaskService {
	return &TaskService{
		repository: repo,
	}
}

func (ts *TaskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*storage.Task, error) {
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}
	
	task := &storage.Task{
		Name:   req.Name,
		URL:    req.URL,
		Type:   req.Type,
		Config: string(configJSON),
	}
	
	if err := ts.repository.CreateTask(ctx, task); err != nil {
		return nil, err
	}
	
	return task, nil
}

func (ts *TaskService) GetTask(ctx context.Context, id int64) (*storage.Task, error) {
	return ts.repository.GetTask(ctx, id)
}

func (ts *TaskService) ListTasks(ctx context.Context, limit, offset int) ([]*storage.Task, error) {
	return ts.repository.ListTasks(ctx, limit, offset)
}

func (ts *TaskService) UpdateTask(ctx context.Context, id int64, req *UpdateTaskRequest) (*storage.Task, error) {
	task, err := ts.repository.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}
	
	if req.Name != nil {
		task.Name = *req.Name
	}
	if req.URL != nil {
		task.URL = *req.URL
	}
	if req.Config != nil {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal config: %w", err)
		}
		task.Config = string(configJSON)
	}
	
	if err := ts.repository.UpdateTask(ctx, task); err != nil {
		return nil, err
	}
	
	return task, nil
}

func (ts *TaskService) DeleteTask(ctx context.Context, id int64) error {
	return ts.repository.DeleteTask(ctx, id)
}

type CreateTaskRequest struct {
	Name   string                 `json:"name" binding:"required"`
	URL    string                 `json:"url" binding:"required,url"`
	Type   string                 `json:"type" binding:"required,oneof=api web"`
	Config map[string]interface{} `json:"config,omitempty"`
}

type UpdateTaskRequest struct {
	Name   *string                `json:"name,omitempty"`
	URL    *string                `json:"url,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

