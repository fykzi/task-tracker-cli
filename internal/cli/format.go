package cli

import (
	"github.com/fatih/color"
	"github.com/fykzi/task-tracker-cli/internal/status"
)

func SprintStatus(s status.Status) string {
	switch s {
	case status.StatusTodo:
		return color.RGB(66, 170, 255).Sprint("📋 todo")
	case status.StatusInProgress:
		return color.RGB(255, 255, 0).Sprint("⏳ in-progress")
	case status.StatusDone:
		return color.RGB(0, 255, 100).Sprint("✅ done")
	default:
		return "unknown"
	}
}
