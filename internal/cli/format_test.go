package cli

import (
	"testing"

	"github.com/fykzi/task-tracker-cli/internal/status"
)

func TestSprintStatus_Valid(t *testing.T) {
	for _, st := range status.ListStatus {
		if SprintStatus(st) == "unknown" {
			t.Errorf("SprintStatus(%q) returned unknown", st)
		}
	}
}

func TestSprintStatus_Unknown(t *testing.T) {
	got := SprintStatus("unknown")

	if got != "unknown" {
		t.Errorf("expected 'unknown', got %q", got)
	}
}
