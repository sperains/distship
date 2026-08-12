package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestTargetIDsAreSorted(t *testing.T) {
	cfg := &Config{Projects: map[string]Project{
		"z": {Environments: map[string]Environment{"prod": {}}},
		"a": {Environments: map[string]Environment{"test": {}, "dev": {}}},
	}}
	want := []string{"a:dev", "a:test", "z:prod"}
	if got := cfg.TargetIDs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("TargetIDs() = %#v, want %#v", got, want)
	}
}

func TestArchiveRenamesConfigurationAndAvoidsCollision(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "projects.toml")
	if err := os.WriteFile(path, []byte("version = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	timestamp := time.Date(2026, 8, 12, 12, 30, 0, 0, time.Local)
	first := path + ".bak-20260812T123000"
	if err := os.WriteFile(first, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	backup, err := Archive(path, timestamp)
	if err != nil {
		t.Fatal(err)
	}
	if backup != first+"-1" {
		t.Fatalf("backup = %q", backup)
	}
	if _, err := os.Stat(backup); err != nil {
		t.Fatalf("backup stat error = %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source still exists or unexpected error: %v", err)
	}
}
