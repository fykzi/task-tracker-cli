package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fykzi/task-tracker-cli/internal/cli"
	"github.com/fykzi/task-tracker-cli/internal/storage"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	sp, err := storagePath()
	if err != nil {
		return err
	}

	s, err := storage.LoadStorage(sp)
	if err != nil {
		return err
	}

	c := cli.New(s)
	return c.Run(os.Args[1:])
}

func storagePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(configDir, "task-tracker-cli")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}

	return filepath.Join(dir, "storage.json"), nil
}
