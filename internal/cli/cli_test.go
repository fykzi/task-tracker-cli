package cli

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fykzi/task-tracker-cli/internal/status"
	"github.com/fykzi/task-tracker-cli/internal/storage"
)

type mockStorage struct {
	tasks  []*storage.Task
	nextID int

	addErr          error
	updateErr       error
	deleteErr       error
	changeStatusErr error
}

func newMockStorage() *mockStorage {
	return &mockStorage{nextID: 1}
}

func (m *mockStorage) Add(description string) (*storage.Task, error) {
	if m.addErr != nil {
		return nil, m.addErr
	}
	now := time.Now()
	task := &storage.Task{
		ID:          m.nextID,
		Description: description,
		Status:      status.StatusTodo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	m.tasks = append(m.tasks, task)
	m.nextID++

	return task, nil
}

func (m *mockStorage) Update(id int, description string) (*storage.Task, string, error) {
	if m.updateErr != nil {
		return nil, "", m.updateErr
	}

	for _, t := range m.tasks {
		if t.ID == id {
			old := t.Description
			t.Description = description
			t.UpdatedAt = time.Now()
			return t, old, nil
		}
	}

	return nil, "", storage.TaskDoesNotExistError{ID: id}
}

func (m *mockStorage) Delete(id int) (*storage.Task, error) {
	if m.deleteErr != nil {
		return nil, m.deleteErr
	}

	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return t, nil
		}
	}

	return nil, storage.TaskDoesNotExistError{ID: id}
}

func (m *mockStorage) ChangeStatus(id int, st status.Status) (*storage.Task, error) {
	if m.changeStatusErr != nil {
		return nil, m.changeStatusErr
	}

	for _, t := range m.tasks {
		if t.ID == id {
			t.Status = st
			t.UpdatedAt = time.Now()

			return t, nil
		}
	}

	return nil, storage.TaskDoesNotExistError{ID: id}
}

func (m *mockStorage) ListTasks() map[status.Status][]*storage.Task {
	result := make(map[status.Status][]*storage.Task)

	for _, t := range m.tasks {
		result[t.Status] = append(result[t.Status], t)
	}

	return result
}

func (m *mockStorage) ListTasksByStatus(st status.Status) []*storage.Task {
	var result []*storage.Task

	for _, t := range m.tasks {
		if t.Status == st {
			result = append(result, t)
		}
	}

	return result
}

func TestWrongUsageError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  WrongUsageError
		want string
	}{
		{
			name: "message only",
			err:  WrongUsageError{Message: "subcommand requied"},
			want: "wrong usage: subcommand requied",
		},
		{
			name: "cause only",
			err: WrongUsageError{
				Cause: TooFewArgumentsError{Need: 1, Have: 0},
			},
			want: "wrong usage: too few arguments: need 1, have 0",
		},
		{
			name: "message and cause",
			err: WrongUsageError{
				Message: "add",
				Cause:   TooFewArgumentsError{Need: 1, Have: 0},
			},
			want: "wrong usage: add: too few arguments: need 1, have 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTooFewArgumentsError_Error(t *testing.T) {
	err := TooFewArgumentsError{Need: 1, Have: 0}
	want := "too few arguments: need 1, have 0"

	if got := err.Error(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRun_NoArgs(t *testing.T) {
	c := New(newMockStorage())

	err := c.Run([]string{})
	if err == nil {
		t.Fatal("expected error  for no args")
	}

	var target WrongUsageError
	if !errors.As(err, &target) {
		t.Errorf("expected WrongUsageError, got %T", err)
	}
}

func TestRun_Help(t *testing.T) {
	c := New(newMockStorage())

	err := c.Run([]string{"help"})
	if err != nil {
		t.Errorf("help should not be return error: %v", err)
	}

	err = c.Run([]string{"-h"})
	if err != nil {
		t.Errorf("-h should not be return error: %v", err)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	c := New(newMockStorage())

	err := c.Run([]string{"unknown"})
	if err == nil {
		t.Fatal("expected error")
	}

	var target WrongUsageError
	if !errors.As(err, &target) {
		t.Errorf("expected WrongUsageError, got %T", err)
	}
}

func TestAddTask(t *testing.T) {
	m := newMockStorage()
	c := New(m)

	err := c.addTask([]string{"valid description"})
	if err != nil {
		t.Fatalf("addTask: %v", err)
	}

	if len(m.tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(m.tasks))
	}
	if m.tasks[0].Description != "valid description" {
		t.Errorf("unexpected description: %q", m.tasks[0].Description)
	}
}

func TestAddTask_NoArgs(t *testing.T) {
	c := New(newMockStorage())

	err := c.addTask([]string{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddTask_EmptyDescription(t *testing.T) {
	c := New(newMockStorage())

	err := c.Run([]string{""})
	if err == nil {
		t.Fatal("expected error for empty description")
	}
}

func TestAddTask_StorageError(t *testing.T) {
	m := newMockStorage()
	m.addErr = errors.New("disk full")
	c := New(m)

	err := c.addTask([]string{"valid description"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseID_Valid(t *testing.T) {
	c := New(newMockStorage())

	id, rest, err := c.parseID([]string{"5", "extra"}, 1, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 5 {
		t.Errorf("expected id=5, got %d", id)
	}
	if len(rest) != 1 || rest[0] != "extra" {
		t.Errorf("unexpected rest: %v", rest)
	}
}

func TestParseID_NotANumber(t *testing.T) {
	c := New(newMockStorage())

	_, _, err := c.parseID([]string{"abc"}, 1, "test")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseID_TooFew(t *testing.T) {
	c := New(newMockStorage())

	_, _, err := c.parseID([]string{}, 1, "test")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListTasks_InvalidStatus(t *testing.T) {
	c := New(newMockStorage())

	err := c.listTasks([]string{"invalid-status"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListTasks_TooManyArgs(t *testing.T) {
	c := New(newMockStorage())

	err := c.listTasks([]string{"todo", "done"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListTasks_Valid(t *testing.T) {
	m := newMockStorage()
	c := New(m)

	_, _ = m.Add("task 1")
	_, _ = m.Add("task 2")

	err := c.listTasks([]string{})
	if err != nil {
		t.Errorf("ListTasks: %v", err)
	}

	err = c.listTasks([]string{"todo"})
	if err != nil {
		t.Errorf("ListTasks todo: %v", err)
	}
}

func TestUpdateTask(t *testing.T) {
	m := newMockStorage()
	c := New(m)

	_, _ = m.Add("original")

	err := c.updateTask([]string{"1", "new description"})
	if err != nil {
		t.Fatalf("updateTask: %v", err)
	}

	if len(m.tasks) != 1 || m.tasks[0].Description != "new description" {
		t.Error("unexpected result")
	}
}

func TestDeleteTask(t *testing.T) {
	m := newMockStorage()
	c := New(m)

	_, _ = m.Add("to delete")

	err := c.deleteTask([]string{"1"})
	if err != nil {
		t.Fatalf("deleteTask: %v", err)
	}

	if len(m.tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(m.tasks))
	}
}

func TestMarkTask(t *testing.T) {
	m := newMockStorage()
	c := New(m)

	task, _ := m.Add("task")
	fmt.Println(c.parseID([]string{"1"}, 1, "mark"))
	err := c.markTask([]string{"1"}, status.StatusDone)
	if err != nil {
		t.Fatalf("markTask: %v", err)
	}

	if m.tasks[0].Status != status.StatusDone {
		t.Errorf("expected done, got %q", m.tasks[0].Status)
	}

	_ = task
}
