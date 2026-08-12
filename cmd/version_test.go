package cmd

import (
	"runtime/debug"
	"testing"
)

func TestResolveBuildMetadataUsesLinkedValues(t *testing.T) {
	info := testBuildInfo()
	metadata := resolveBuildMetadata("0.1.0", "linked-commit", "linked-date", info)

	if metadata.version != "0.1.0" || metadata.commit != "linked-commit" || metadata.date != "linked-date" {
		t.Fatalf("metadata = %+v, want linked values", metadata)
	}
}

func TestResolveBuildMetadataFallsBackToGoBuildInfo(t *testing.T) {
	metadata := resolveBuildMetadata("dev", "none", "0001-01-01T00:00:00Z", testBuildInfo())

	if metadata.version != "v0.1.0" {
		t.Fatalf("version = %q, want v0.1.0", metadata.version)
	}
	if metadata.commit != "0123456789abcdef" {
		t.Fatalf("commit = %q, want VCS revision", metadata.commit)
	}
	if metadata.date != "2026-08-12T00:00:00Z" {
		t.Fatalf("date = %q, want VCS time", metadata.date)
	}
}

func TestResolveBuildMetadataIgnoresDevelopmentModuleVersion(t *testing.T) {
	info := testBuildInfo()
	info.Main.Version = "(devel)"
	metadata := resolveBuildMetadata("dev", "none", "unknown", info)

	if metadata.version != "dev" {
		t.Fatalf("version = %q, want dev", metadata.version)
	}
}

func testBuildInfo() *debug.BuildInfo {
	return &debug.BuildInfo{
		Main: debug.Module{Version: "v0.1.0"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef"},
			{Key: "vcs.time", Value: "2026-08-12T00:00:00Z"},
		},
	}
}
