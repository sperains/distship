package history

import (
	"os"
	"path/filepath"
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
}
