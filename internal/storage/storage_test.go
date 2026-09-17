package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/fykzi/task-tracker-cli/internal/status"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "storage.json")

	s, err := LoadStorage(path)
	if err != nil {
		t.Fatalf("LoadStorage: %v", err)
	}

	return s
}

func newTestStorageWithPath(t *testing.T) (*Storage, string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "storage.json")

	s, err := LoadStorage(path)
	if err != nil {
		t.Fatalf("LoadStorage: %v", err)
	}

	return s, path
}

func TestLoadStorage_NewFile(t *testing.T) {
	s := newTestStorage(t)

	if len(s.data.Tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(s.data.Tasks))
	}
	if s.data.NextTaskID != 1 {
		t.Errorf("expected NextTaskID=1, got %d", s.data.NextTaskID)
	}
}

func TestLoadStorage_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "storage.json")

	original := StorageData{
		Tasks: []*Task{
			{
				ID:          5,
				Description: "existing task",
				Status:      status.StatusDone,
			},
		},
		NextTaskID: 6,
	}

	data, _ := json.Marshal(original)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}

	s, err := LoadStorage(path)
	if err != nil {
		t.Fatalf("LoadStorage: %v", err)
	}

	if len(s.data.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(s.data.Tasks))
	}
	if s.data.Tasks[0].ID != 5 {
		t.Errorf("expected ID=5, got %d", s.data.Tasks[0].ID)
	}
	if s.data.NextTaskID != 6 {
		t.Errorf("expected NextTaskID=6, got %d", s.data.NextTaskID)
	}
}

func TestLoadStorage_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "storage.json")

	if err := os.WriteFile(path, []byte("{invalid json}"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadStorage(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadStorage_PreservesPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "storage.json")

	s, err := LoadStorage(path)
	if err != nil {
		t.Fatal(err)
	}

	if s.path != path {
		t.Errorf("expected path %q, got %q", path, s.path)
	}
}

func TestAdd(t *testing.T) {
	s := newTestStorage(t)

	task, err := s.Add("test task")
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	if task.ID != 1 {
		t.Errorf("expeced ID=1, got %d", task.ID)
	}
	if task.Description != "test task" {
		t.Errorf("expected Description='test task', got %q", task.Description)
	}
	if task.Status != status.StatusTodo {
		t.Errorf("expected todo, got %q", task.Status)
	}
	if task.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if task.CreatedAt != task.UpdatedAt {
		t.Error("CreatedAt and UpdatedAt should be equal on creation")
	}
}

func TestAdd_IncrementID(t *testing.T) {
	s := newTestStorage(t)

	t1, _ := s.Add("first")
	t2, _ := s.Add("second")
	t3, _ := s.Add("third")

	if t1.ID != 1 || t2.ID != 2 || t3.ID != 3 {
		t.Errorf("expected IDs 1,2,3, got %d,%d,%d", t1.ID, t2.ID, t3.ID)
	}
}

func TestAdd_PersistToDisk(t *testing.T) {
	s, path := newTestStorageWithPath(t)

	_, err := s.Add("persistent task")
	if err != nil {
		t.Fatal(err)
	}

	s2, err := LoadStorage(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(s2.data.Tasks) != 1 {
		t.Fatalf("expected 1 task after reload, got %d", len(s.data.Tasks))
	}
	if s2.data.Tasks[0].Description != "persistent task" {
		t.Errorf("expected 'persistent task', got %q", s2.data.Tasks[0].Description)
	}
}

func TestAdd_RollsBackOnSaveError(t *testing.T) {
	s := newTestStorage(t)

	s.path = "/nonexistent/storage.json"

	_, err := s.Add("should fail")
	if err == nil {
		t.Fatal("expected save error")
	}

	if len(s.data.Tasks) != 0 {
		t.Errorf("expected 0 tasks in memory, got %d", len(s.data.Tasks))
	}
	if s.data.NextTaskID != 1 {
		t.Errorf("expected NextTaskID=1, got %d", s.data.NextTaskID)
	}
}

func TestGetByID_Found(t *testing.T) {
	s := newTestStorage(t)

	added, _ := s.Add("find me")

	got, ok := s.GetByID(added.ID)
	if !ok {
		t.Fatal("task not found")
	}

	if got.ID != added.ID {
		t.Errorf("expected ID=%d, got %d", added.ID, got.ID)
	}
}

func TestBetByID_NotFound(t *testing.T) {
	s := newTestStorage(t)

	_, ok := s.GetByID(99)
	if ok {
		t.Error("expected not found, got found")
	}
}

func TestUpdate(t *testing.T) {
	s := newTestStorage(t)
	task, _ := s.Add("original")

	updated, old, err := s.Update(task.ID, "new description")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if old != "original" {
		t.Errorf("expected old 'original', got %q", old)
	}
	if updated.Description != "new description" {
		t.Errorf("expected 'new description', got %q", updated.Description)
	}
	if updated.UpdatedAt.Before(updated.CreatedAt) {
		t.Error("UpdatedAt should be after CreatedAt")
	}
}

func TestUpdated_NoChange(t *testing.T) {
	s := newTestStorage(t)
	task, _ := s.Add("same")

	updated, old, err := s.Update(task.ID, "same")
	if err != nil {
		t.Fatal(err)
	}

	if old != "same" || updated.Description != "same" {
		t.Error("expected same description")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	s := newTestStorage(t)

	_, _, err := s.Update(99, "new")
	if err == nil {
		t.Fatal("expected error")
	}

	var target TaskDoesNotExistError

	if !errors.As(err, &target) {
		t.Errorf("expected TaskDoesNotExistError, got %T", err)
	}
	if target.ID != 99 {
		t.Errorf("expected ID=99, got %d", target.ID)
	}
}

func TestUpdate_RollsBackSaveError(t *testing.T) {
	s := newTestStorage(t)
	task, _ := s.Add("original")

	s.path = "/nonexistent/storage.json"

	_, _, err := s.Update(task.ID, "should not apply")
	if err == nil {
		t.Fatal("expected save error")
	}

	got, _ := s.GetByID(task.ID)
	if got.Description != "original" {
		t.Errorf("expected 'original' in memory, got %q", got.Description)
	}
}

func TestDelete(t *testing.T) {
	s := newTestStorage(t)
	task, _ := s.Add("to delete")

	removed, err := s.Delete(task.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if removed.ID != task.ID {
		t.Errorf("expected removed ID=%d, got %d", removed.ID, task.ID)
	}
	if _, ok := s.GetByID(task.ID); ok {
		t.Error("task should be deleted")
	}
}

func TestDelete_NotFound(t *testing.T) {
	s := newTestStorage(t)

	_, err := s.Delete(99)
	if err == nil {
		t.Fatal("expted error")
	}
}

func TestDelete_RollsBackOnSaveError(t *testing.T) {
	s := newTestStorage(t)
	task, _ := s.Add("keep me")

	s.path = "/nonexistent/storage.json"

	_, err := s.Delete(task.ID)
	if err == nil {
		t.Fatal("expected save error")
	}

	if _, ok := s.GetByID(task.ID); !ok {
		t.Error("task should be NOT be deleted from memory")
	}
}

func TestDelete_MultipleTasks(t *testing.T) {
	s := newTestStorage(t)
	t1, _ := s.Add("first")
	t2, _ := s.Add("second")
	t3, _ := s.Add("third")

	_, err := s.Delete(t2.ID)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := s.GetByID(t1.ID); !ok {
		t.Error("t1 should still exist")
	}
	if _, ok := s.GetByID(t2.ID); ok {
		t.Error("t2 should be deleted")
	}
	if _, ok := s.GetByID(t3.ID); !ok {
		t.Error("t3 should still exist")
	}
}

func TestChangeStatus(t *testing.T) {
	s := newTestStorage(t)
	task, _ := s.Add("task")

	updated, err := s.ChangeStatus(task.ID, status.StatusInProgress)
	if err != nil {
		t.Fatalf("ChangeStatus: %v", err)
	}

	if updated.Status != status.StatusInProgress {
		t.Errorf("expected in-progress, got %q", updated.Status)
	}
}

func TestChangeStatus_SameStatus(t *testing.T) {
	s := newTestStorage(t)
	task, _ := s.Add("task")

	updated, err := s.ChangeStatus(task.ID, status.StatusTodo)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != status.StatusTodo {
		t.Errorf("expected todo, got %q", updated.Status)
	}
}

func TestChangeStatus_NotFound(t *testing.T) {
	s := newTestStorage(t)

	_, err := s.ChangeStatus(99, status.StatusInProgress)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestChange_RollsBackOnSaveError(t *testing.T) {
	s := newTestStorage(t)
	task, _ := s.Add("task")

	s.path = "/nonexistent/storage.json"

	_, err := s.ChangeStatus(task.ID, status.StatusDone)
	if err == nil {
		t.Fatal("expected save error")
	}

	got, _ := s.GetByID(task.ID)
	if got.Status != status.StatusTodo {
		t.Errorf("expected todo in memory, got %q", got.Status)
	}
}

func TestListTasks(t *testing.T) {
	s := newTestStorage(t)

	t1, _ := s.Add("todo 1")
	t2, _ := s.Add("todo 2")
	t3, _ := s.Add("in progress")
	t4, _ := s.Add("done")

	_, _ = s.ChangeStatus(t3.ID, status.StatusInProgress)
	_, _ = s.ChangeStatus(t4.ID, status.StatusDone)

	grouped := s.ListTasks()

	if len(grouped[status.StatusTodo]) != 2 {
		t.Errorf("expected 2 todo, got %d", len(grouped[status.StatusTodo]))
	}
	if len(grouped[status.StatusInProgress]) != 1 {
		t.Errorf("expected 1 in-progress, got %d", len(grouped[status.StatusInProgress]))
	}
	if len(grouped[status.StatusDone]) != 1 {
		t.Errorf("expected 1 done, got %d", len(grouped[status.StatusDone]))
	}

	_ = t1
	_ = t2
}

func TestListTasksByStatus(t *testing.T) {
	s := newTestStorage(t)

	_, _ = s.Add("todo")
	t2, _ := s.Add("done")

	_, _ = s.ChangeStatus(t2.ID, status.StatusDone)

	todos := s.ListTasksByStatus(status.StatusTodo)
	if len(todos) != 1 {
		t.Errorf("expected 1 todo, got %d", len(todos))
	}

	dones := s.ListTasksByStatus(status.StatusDone)
	if len(dones) != 1 {
		t.Errorf("expected 1 done, got %d", len(dones))
	}

	empty := s.ListTasksByStatus(status.StatusInProgress)
	if len(empty) != 0 {
		t.Errorf("expected 0 in-progress, got %d", len(empty))
	}
}
