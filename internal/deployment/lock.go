package deployment

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type Lock struct {
	path string
}

func Acquire(stateDirectory, targetID string) (*Lock, error) {
	directory := filepath.Join(stateDirectory, "locks")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create deployment lock directory: %w", err)
	}
	name := strings.ReplaceAll(targetID, ":", "--") + ".lock"
	path := filepath.Join(directory, name)
	file, err := createLock(path)
	if errors.Is(err, os.ErrExist) {
		stale, staleErr := staleLock(path)
		if staleErr != nil {
			return nil, fmt.Errorf("inspect deployment lock %s: %w", path, staleErr)
		}
		if stale {
			if removeErr := os.Remove(path); removeErr != nil {
				return nil, fmt.Errorf("remove stale deployment lock %s: %w", path, removeErr)
			}
			file, err = createLock(path)
		}
	}
	if errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("deployment is already running for %s (lock: %s)", targetID, path)
	}
	if err != nil {
		return nil, fmt.Errorf("create deployment lock: %w", err)
	}
	if _, err := fmt.Fprintf(file, "%d\n", os.Getpid()); err != nil {
		file.Close()
		os.Remove(path)
		return nil, fmt.Errorf("write deployment lock: %w", err)
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return nil, fmt.Errorf("close deployment lock: %w", err)
	}
	return &Lock{path: path}, nil
}

func createLock(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
}

func staleLock(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return false, fmt.Errorf("invalid process ID")
	}
	err = syscall.Kill(pid, 0)
	if errors.Is(err, syscall.ESRCH) {
		return true, nil
	}
	return false, nil
}

func (l *Lock) Release() error {
	if l == nil || l.path == "" {
		return nil
	}
	err := os.Remove(l.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
