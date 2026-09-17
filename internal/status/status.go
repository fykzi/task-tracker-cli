package status

import (
	"fmt"
	"slices"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusDone       Status = "done"
)

var ListStatus = []Status{StatusTodo, StatusInProgress, StatusDone}

func IsValid(status Status) bool {
	return slices.Contains(ListStatus, status)
}

func Parse(s string) (Status, error) {
	if s == "" {
		return "", fmt.Errorf("empty status")
	}

	status := Status(s)

	if !IsValid(status) {
		return "", fmt.Errorf("invalid status: %q", s)
	}

	return status, nil
}
