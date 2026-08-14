package history

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sperains/distship/internal/config"
)

func TestAppendAndFindLastMatchingTarget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state", "history.jsonl")
	ref := config.TargetRef{ProjectID: "site", EnvironmentID: "test"}
	target := config.Target{Host: "server", Directory: "/www/site"}
	first := Record{DeployedAt: time.Date(2026, 8, 11, 9, 0, 0, 0, time.UTC), Project: "site", Environment: "test", Host: "server", Directory: "/www/site", Commit: "first"}
	other := Record{DeployedAt: time.Date(2026, 8, 11, 10, 0, 0, 0, time.UTC), Project: "other", Environment: "test", Host: "server", Directory: "/www/site", Commit: "other"}
	latest := Record{DeployedAt: time.Date(2026, 8, 12, 9, 0, 0, 0, time.UTC), Project: "site", Environment: "test", Host: "server", Directory: "/www/site", Commit: "latest"}
	for _, record := range []Record{first, other, latest} {
		if err := Append(path, record); err != nil {
			t.Fatal(err)
		}
	}

	record, found, err := Last(path, ref, target)
	if err != nil || !found || record.Commit != "latest" {
		t.Fatalf("record = %#v, found = %v, error = %v", record, found, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("permissions = %o", info.Mode().Perm())
	}
}

func TestLastAllowsMissingHistory(t *testing.T) {
	_, found, err := Last(filepath.Join(t.TempDir(), "missing.jsonl"), config.TargetRef{}, config.Target{})
	if err != nil || found {
		t.Fatalf("found = %v, error = %v", found, err)
	}
}

func TestLastRejectsMalformedHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	if err := os.WriteFile(path, []byte("not-json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _, err := Last(path, config.TargetRef{}, config.Target{})
	if err == nil {
		t.Fatal("Last() error = nil")
	}
	if !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "line 1") {
		t.Fatalf("Last() error = %v, want path and line", err)
	}
}

func TestListReportsUnsupportedVersionLocation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	if err := os.WriteFile(path, []byte("{\"version\":2}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := List(path, nil, 10)
	if err == nil || !strings.Contains(err.Error(), path) || !strings.Contains(err.Error(), "line 1") || !strings.Contains(err.Error(), "version 2") {
		t.Fatalf("List() error = %v", err)
	}
}

func TestListFiltersLimitsAndReturnsLatestFirst(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	records := []Record{
		{Project: "site", Environment: "test", Commit: "first"},
		{Project: "other", Environment: "test", Commit: "other"},
		{Project: "site", Environment: "test", Commit: "second"},
		{Project: "site", Environment: "test", Commit: "latest"},
	}
	for _, record := range records {
		if err := Append(path, record); err != nil {
			t.Fatal(err)
		}
	}
	ref := config.TargetRef{ProjectID: "site", EnvironmentID: "test"}
	listed, err := List(path, &ref, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 2 || listed[0].Commit != "latest" || listed[1].Commit != "second" {
		t.Fatalf("List() = %#v", listed)
	}
}

func TestListAllowsMissingHistory(t *testing.T) {
	records, err := List(filepath.Join(t.TempDir(), "missing.jsonl"), nil, 10)
	if err != nil || len(records) != 0 {
		t.Fatalf("records = %#v, error = %v", records, err)
	}
}

func TestListRejectsInvalidLimit(t *testing.T) {
	if _, err := List(filepath.Join(t.TempDir(), "history.jsonl"), nil, 0); err == nil {
		t.Fatal("List() error = nil")
	}
}
