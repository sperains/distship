package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sperains/distship/internal/config"
)

const Version = 1

type Record struct {
	Version       int       `json:"version"`
	DeployedAt    time.Time `json:"deployedAt"`
	Project       string    `json:"project"`
	Environment   string    `json:"environment"`
	Host          string    `json:"host"`
	Directory     string    `json:"directory"`
	Branch        string    `json:"branch,omitempty"`
	Commit        string    `json:"commit,omitempty"`
	Dirty         bool      `json:"dirty,omitempty"`
	Artifact      string    `json:"artifact"`
	DurationMilli int64     `json:"durationMs"`
}

func DefaultPath() (string, error) {
	if stateHome := os.Getenv("XDG_STATE_HOME"); stateHome != "" {
		return filepath.Join(stateHome, "distship", "history.jsonl"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determine state directory: %w", err)
	}
	return filepath.Join(home, ".local", "state", "distship", "history.jsonl"), nil
}

func Last(path string, ref config.TargetRef, target config.Target) (Record, bool, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, fmt.Errorf("open deployment history: %w", err)
	}
	defer file.Close()

	var last Record
	found := false
	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)
	for scanner.Scan() {
		var record Record
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return Record{}, false, fmt.Errorf("parse deployment history: %w", err)
		}
		if record.Version != Version {
			return Record{}, false, fmt.Errorf("unsupported deployment history version: %d", record.Version)
		}
		if record.Project == ref.ProjectID && record.Environment == ref.EnvironmentID && record.Host == target.Host && record.Directory == target.Directory {
			last = record
			found = true
		}
	}
	if err := scanner.Err(); err != nil {
		return Record{}, false, fmt.Errorf("read deployment history: %w", err)
	}
	return last, found, nil
}

func Append(path string, record Record) error {
	if record.Version == 0 {
		record.Version = Version
	}
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("encode deployment history: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open deployment history: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return fmt.Errorf("set deployment history permissions: %w", err)
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		file.Close()
		return fmt.Errorf("append deployment history: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync deployment history: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close deployment history: %w", err)
	}
	return nil
}
