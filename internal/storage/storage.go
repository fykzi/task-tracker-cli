package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/fykzi/task-tracker-cli/internal/status"
)

type TaskDoesNotExistError struct {
	ID int
}

type Task struct {
	ID          int           `json:"id"`
	Description string        `json:"description"`
	Status      status.Status `json:"status"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StorageData struct {
	Tasks      []*Task `json:"tasks"`
	NextTaskID int     `json:"next_task_id"`
}

type Storage struct {
	data StorageData
	path string
}

func (e TaskDoesNotExistError) Error() string {
	return fmt.Sprintf("task with ID(%d) does not exist", e.ID)
}

func (t *Task) String() string {
	return fmt.Sprintf("Task{ID: %d, Description: %s, Status: %s}", t.ID, t.Description, t.Status)
}

func LoadStorage(path string) (*Storage, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Storage{
			data: StorageData{
				Tasks:      []*Task{},
				NextTaskID: 1,
			},
			path: path,
		}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read storage file: %w", err)
	}

	var storage Storage

	if err = json.Unmarshal(data, &storage.data); err != nil {
		return nil, fmt.Errorf("parse storage file: %w", err)
	}

	storage.path = path
	return &storage, nil
}

func (s *Storage) commit(mutate func(StorageData) StorageData) error {
	if mutate == nil {
		return fmt.Errorf("commit: nil mutate function")
	}
	newData := mutate(s.data)

	snapshot := &Storage{data: newData, path: s.path}
	if err := snapshot.save(); err != nil {
		return err
	}

	s.data = newData
	return nil
}

func (s *Storage) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)

		return fmt.Errorf("save: %w", err)
	}

	return nil
}

func (s *Storage) Add(description string) (*Task, error) {
	now := time.Now()
	newTask := &Task{
		ID:          s.data.NextTaskID,
		Description: description,
		Status:      status.StatusTodo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.commit(func(d StorageData) StorageData {
		newTasks := make([]*Task, len(d.Tasks), len(d.Tasks)+1)
		copy(newTasks, d.Tasks)
		newTasks = append(newTasks, newTask)

		return StorageData{Tasks: newTasks, NextTaskID: d.NextTaskID + 1}
	})
	if err != nil {
		return nil, err
	}

	return newTask, nil
}

func (s *Storage) GetByID(id int) (*Task, bool) {
	for _, task := range s.data.Tasks {
		if task.ID == id {
			return task, true
		}
	}

	return nil, false
}

func (s *Storage) Update(id int, description string) (*Task, string, error) {
	task, ok := s.GetByID(id)
	if !ok {
		return nil, "", TaskDoesNotExistError{ID: id}
	}

	if task.Description == description {
		return task, description, nil
	}

	oldDescription := task.Description

	updated := *task
	updated.Description = description
	updated.UpdatedAt = time.Now()

	err := s.commit(func(d StorageData) StorageData {
		return StorageData{
			Tasks:      replaceTask(d.Tasks, &updated),
			NextTaskID: d.NextTaskID,
		}
	})
	if err != nil {
		return nil, "", err
	}

	return &updated, oldDescription, nil
}

func (s *Storage) Delete(id int) (*Task, error) {
	idx := -1

	for i, task := range s.data.Tasks {
		if task.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return nil, TaskDoesNotExistError{ID: id}
	}

	removed := s.data.Tasks[idx]

	err := s.commit(func(d StorageData) StorageData {
		newTasks := make([]*Task, 0, len(d.Tasks)-1)
		newTasks = append(newTasks, d.Tasks[:idx]...)
		newTasks = append(newTasks, d.Tasks[idx+1:]...)

		return StorageData{Tasks: newTasks, NextTaskID: d.NextTaskID}
	})
	if err != nil {
		return nil, err
	}

	return removed, nil
}

func (s *Storage) ListTasks() map[status.Status][]*Task {
	result := make(map[status.Status][]*Task)

	for _, task := range s.data.Tasks {
		result[task.Status] = append(result[task.Status], task)
	}

	return result
}

func (s *Storage) ListTasksByStatus(st status.Status) []*Task {
	var result []*Task

	for _, t := range s.data.Tasks {
		if t.Status == st {
			result = append(result, t)
		}
	}

	return result
}

func (s *Storage) ChangeStatus(id int, status status.Status) (*Task, error) {
	task, ok := s.GetByID(id)
	if !ok {
		return nil, TaskDoesNotExistError{ID: id}
	}

	if task.Status == status {
		return task, nil
	}

	updated := *task
	updated.Status = status
	updated.UpdatedAt = time.Now()

	err := s.commit(func(d StorageData) StorageData {
		return StorageData{
			Tasks:      replaceTask(d.Tasks, &updated),
			NextTaskID: d.NextTaskID,
		}
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func replaceTask(tasks []*Task, replacement *Task) []*Task {
	newTasks := make([]*Task, len(tasks))
	copy(newTasks, tasks)

	for i, t := range newTasks {
		if t.ID == replacement.ID {
			newTasks[i] = replacement
			break
		}
	}
	return newTasks
}
