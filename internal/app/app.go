package app

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"zan/internal/store"
	"zan/internal/task"

	"github.com/google/uuid"
)

// App はアプリケーションの主要なロジックを管理します。
type App struct {
	Tasks *task.Tasks
}

// NewApp は新しいAppインスタンスを作成し、タスクデータをロードします。
func NewApp() (*App, error) {
	tasks, err := store.LoadTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	// If no tasks are loaded and not in test environment, add some dummy data for demonstration
	if len(tasks.Tasks) == 0 && os.Getenv("ZAN_TEST_ENV") != "true" {
		now := time.Now()
		tasks.Tasks = []task.Task{
			{
				ID:          uuid.New().String(),
				Title:       "Buy groceries",
				Description: "Milk, eggs, bread, and cheese",
				Status:      task.StatusTODO,
				Priority:    task.PriorityHigh,
				Tags:        []string{"personal", "urgent"},
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          uuid.New().String(),
				Title:       "Finish report",
				Description: "Complete the Q3 financial report",
				Status:      task.StatusInProgress,
				Priority:    task.PriorityMedium,
				Tags:        []string{"work"},
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          uuid.New().String(),
				Title:       "Call John",
				Description: "Discuss project updates",
				Status:      task.StatusPending,
				Priority:    task.PriorityLow,
				Tags:        []string{"personal"},
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}
		// Save dummy data if auto-save is enabled
		if tasks.Settings.AutoSave {
			if err := store.SaveTasks(tasks); err != nil {
				return nil, fmt.Errorf("failed to auto-save dummy tasks: %w", err)
			}
		}
	}

	return &App{Tasks: tasks}, nil
}

// AddTask は新しいタスクを作成し、タスクリストに追加します。
func (a *App) AddTask(title, description string, priority task.Priority, tags []string) (*task.Task, error) {
	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	if priority == "" {
		priority = a.Tasks.Settings.DefaultPriority
	}

	newTask := task.Task{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Status:      task.StatusTODO,
		Priority:    priority,
		Tags:        tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := newTask.Validate(); err != nil {
		return nil, fmt.Errorf("invalid task data: %w", err)
	}

	a.Tasks.Tasks = append(a.Tasks.Tasks, newTask)
	if a.Tasks.Settings.AutoSave {
		if err := store.SaveTasks(a.Tasks); err != nil {
			return nil, fmt.Errorf("failed to auto-save tasks: %w", err)
		}
	}
	return &newTask, nil
}

// GetTaskByID は指定されたIDのタスクを返します。
func (a *App) GetTaskByID(id string) (*task.Task, error) {
	for _, t := range a.Tasks.Tasks {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("task with ID %s not found", id)
}

// UpdateTask は既存のタスクを更新します。
func (a *App) UpdateTask(id, title, description string, status task.Status, priority task.Priority, tags []string) (*task.Task, error) {
	for i, t := range a.Tasks.Tasks {
		if t.ID == id {
			if title != "" {
				a.Tasks.Tasks[i].Title = title
			}
			if description != "" {
				a.Tasks.Tasks[i].Description = description
			}
			if status != "" {
				a.Tasks.Tasks[i].Status = status
				if status == task.StatusDone {
					now := time.Now()
					a.Tasks.Tasks[i].CompletedAt = &now
				} else {
					a.Tasks.Tasks[i].CompletedAt = nil
				}
			}
			if priority != "" {
				a.Tasks.Tasks[i].Priority = priority
			}
			if tags != nil {
				a.Tasks.Tasks[i].Tags = tags
			}
			a.Tasks.Tasks[i].UpdatedAt = time.Now()

			if err := a.Tasks.Tasks[i].Validate(); err != nil {
				return nil, fmt.Errorf("invalid task data after update: %w", err)
			}

			if a.Tasks.Settings.AutoSave {
				if err := store.SaveTasks(a.Tasks); err != nil {
					return nil, fmt.Errorf("failed to auto-save tasks: %w", err)
				}
			}
			return &a.Tasks.Tasks[i], nil
		}
	}
	return nil, fmt.Errorf("task with ID %s not found", id)
}

// DeleteTask は指定されたIDのタスクを削除します。
func (a *App) DeleteTask(id string) error {
	for i, t := range a.Tasks.Tasks {
		if t.ID == id {
			a.Tasks.Tasks = append(a.Tasks.Tasks[:i], a.Tasks.Tasks[i+1:]...)
			if a.Tasks.Settings.AutoSave {
				if err := store.SaveTasks(a.Tasks); err != nil {
					return fmt.Errorf("failed to auto-save tasks after deletion: %w", err)
				}
			}
			return nil
		}
	}
	return fmt.Errorf("task with ID %s not found", id)
}

// GetAllTasks は全てのタスクを返します。
func (a *App) GetAllTasks() []task.Task {
	return a.Tasks.Tasks
}

// GetTaskStats はタスクの統計情報を返します。
func (a *App) GetTaskStats() (total, completed, incomplete int) {
	total = len(a.Tasks.Tasks)
	for _, t := range a.Tasks.Tasks {
		if t.Status == task.StatusDone {
			completed++
		} else {
			incomplete++
		}
	}
	return
}

// Search はキーワードに基づいてタスクを検索します。
func (a *App) Search(keyword string) []task.Task {
	return task.SearchTasks(a.Tasks.Tasks, keyword)
}

// GetFilteredTasksByStatus は指定されたステータスでタスクをフィルタリングして返します。
func (a *App) GetFilteredTasksByStatus(statuses []task.Status) []task.Task {
	if len(statuses) == 0 {
		return a.Tasks.Tasks
	}

	var filteredTasks []task.Task
	statusMap := make(map[task.Status]bool)
	for _, s := range statuses {
		statusMap[s] = true
	}

	for _, t := range a.Tasks.Tasks {
		if statusMap[t.Status] {
			filteredTasks = append(filteredTasks, t)
		}
	}
	return filteredTasks
}

// GetFilteredTasksByTags は指定されたタグでタスクをフィルタリングして返します。
// 複数のタグが指定された場合、それら全てのタグを持つタスクを返します (AND検索)。
func (a *App) GetFilteredTasksByTags(tags []string) []task.Task {
	if len(tags) == 0 {
		return a.Tasks.Tasks
	}

	var filteredTasks []task.Task
	for _, t := range a.Tasks.Tasks {
		matchCount := 0
		for _, filterTag := range tags {
			for _, taskTag := range t.Tags {
				if strings.EqualFold(strings.TrimSpace(filterTag), strings.TrimSpace(taskTag)) {
					matchCount++
					break
				}
			}
		}
		if matchCount == len(tags) {
			filteredTasks = append(filteredTasks, t)
		}
	}
	return filteredTasks
}

// SortTasks は指定された基準と順序でタスクをソートします。
func (a *App) SortTasks(tasks []task.Task, sortBy string, ascending bool) []task.Task {
	if len(tasks) == 0 {
		return tasks
	}

	sort.Slice(tasks, func(i, j int) bool {
		switch sortBy {
		case "created_at":
			if ascending {
				return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
			}
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		case "updated_at":
			if ascending {
				return tasks[i].UpdatedAt.Before(tasks[j].UpdatedAt)
			}
			return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
		case "priority":
			// 優先度はHIGH > MEDIUM > LOW の順
			priorityOrder := map[task.Priority]int{
				task.PriorityHigh:   3,
				task.PriorityMedium: 2,
				task.PriorityLow:    1,
			}
			p1 := priorityOrder[tasks[i].Priority]
			p2 := priorityOrder[tasks[j].Priority]
			if ascending {
				return p1 < p2
			}
			return p1 > p2
		default:
			// デフォルトは作成日時で降順
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		}
	})
	return tasks
}

// GetFilteredTasksByPriority は指定された優先度でタスクをフィルタリングして返します。
func (a *App) GetFilteredTasksByPriority(priorities []task.Priority) []task.Task {
	if len(priorities) == 0 {
		return a.Tasks.Tasks
	}

	var filteredTasks []task.Task
	priorityMap := make(map[task.Priority]bool)
	for _, p := range priorities {
		priorityMap[p] = true
	}

	for _, t := range a.Tasks.Tasks {
		if priorityMap[t.Priority] {
			filteredTasks = append(filteredTasks, t)
		}
	}
	return filteredTasks
}

// GetAllUniqueTags は全てのタスクからユニークなタグのリストを返します。
func (a *App) GetAllUniqueTags() []string {
	uniqueTags := make(map[string]bool)
	var tags []string
	for _, t := range a.Tasks.Tasks {
		for _, tag := range t.Tags {
			trimmedTag := strings.TrimSpace(tag)
			if trimmedTag != "" && !uniqueTags[trimmedTag] {
				uniqueTags[trimmedTag] = true
				tags = append(tags, trimmedTag)
			}
		}
	}
	sort.Strings(tags) // タグをアルファベット順にソート
	return tags
}
