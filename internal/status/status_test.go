package status

import (
	"fmt"
	"testing"
)

func TestParse_Valid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Status
	}{
		{"todo", "todo", StatusTodo},
		{"in-progess", "in-progress", StatusInProgress},
		{"done", "done", StatusDone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexprected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("Parse(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParse_Invalid(t *testing.T) {
	tests := []string{
		"",
		"TODO",
		"in progress",
		"DONE",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := Parse(input)
			if err == nil {
				t.Errorf("Parse(%q) expected error, got nil", input)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"todo", StatusTodo, true},
		{"in-progress", StatusInProgress, true},
		{"done", StatusDone, true},
		{"empty", Status(""), false},
		{"unknow", Status("unknow"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.status); got != tt.want {
				fmt.Errorf("IsValid(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
