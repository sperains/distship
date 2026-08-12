package deployment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAcquirePreventsConcurrentDeploymentAndReleases(t *testing.T) {
	directory := t.TempDir()
	first, err := Acquire(directory, "site:test")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Acquire(directory, "site:test"); err == nil {
		t.Fatal("second Acquire() error = nil")
	}
	if err := first.Release(); err != nil {
		t.Fatal(err)
	}
	second, err := Acquire(directory, "site:test")
	if err != nil {
		t.Fatal(err)
	}
	if err := second.Release(); err != nil {
		t.Fatal(err)
	}
}

func TestAcquireReplacesStaleLock(t *testing.T) {
	directory := t.TempDir()
	locks := filepath.Join(directory, "locks")
	if err := os.MkdirAll(locks, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(locks, "site--test.lock"), []byte("99999999\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	lock, err := Acquire(directory, "site:test")
	if err != nil {
		t.Fatal(err)
	}
	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}
}
