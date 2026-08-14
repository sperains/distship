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

const (
	initialScanBuffer = 64 * 1024
	maximumRecordSize = 1024 * 1024
)

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

func (r Record) Duration() time.Duration {
	return time.Duration(r.DurationMilli) * time.Millisecond
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
	var last Record
	found := false
	err := scanRecords(path, func(record Record) {
		if record.matches(ref) && record.Host == target.Host && record.Directory == target.Directory {
			last = record
			found = true
		}
	})
	if err != nil {
		return Record{}, false, err
	}
	return last, found, nil
}

func List(path string, ref *config.TargetRef, limit int) ([]Record, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("deployment history limit must be greater than zero")
	}

	var records []Record
	err := scanRecords(path, func(record Record) {
		if ref != nil && !record.matches(*ref) {
			return
		}
		if len(records) == limit {
			copy(records, records[1:])
			records[len(records)-1] = record
			return
		}
		records = append(records, record)
	})
	if err != nil {
		return nil, err
	}
	for left, right := 0, len(records)-1; left < right; left, right = left+1, right-1 {
		records[left], records[right] = records[right], records[left]
	}
	return records, nil
}

func (r Record) matches(ref config.TargetRef) bool {
	return r.Project == ref.ProjectID && r.Environment == ref.EnvironmentID
}

func scanRecords(path string, visit func(Record)) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open deployment history %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, initialScanBuffer)
	scanner.Buffer(buffer, maximumRecordSize)
	line := 0
	for scanner.Scan() {
		line++
		var record Record
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return fmt.Errorf("parse deployment history %s at line %d: %w", path, line, err)
		}
		if record.Version != Version {
			return fmt.Errorf("parse deployment history %s at line %d: unsupported version %d", path, line, record.Version)
		}
		visit(record)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read deployment history %s at line %d: %w", path, line+1, err)
	}
	return nil
}

func Append(path string, record Record) (resultErr error) {
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
	defer func() {
		if err := file.Close(); resultErr == nil && err != nil {
			resultErr = fmt.Errorf("close deployment history: %w", err)
		}
	}()
	if err := file.Chmod(0o600); err != nil {
		return fmt.Errorf("set deployment history permissions: %w", err)
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("append deployment history: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync deployment history: %w", err)
	}
	return nil
}
