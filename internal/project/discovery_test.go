package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverNodeProjectDefaults(t *testing.T) {
	directory := t.TempDir()
	manifest := []byte(`{"name":"My Web_App","scripts":{"build-test":"vite build --mode test","build":"vite build"}}`)
	if err := os.WriteFile(filepath.Join(directory, "package.json"), manifest, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "pnpm-lock.yaml"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(directory, "dist"), 0o700); err != nil {
		t.Fatal(err)
	}

	defaults := Discover(directory, "test")
	if defaults.ProjectID != "my-web-app" || defaults.ProjectName != "My Web_App" || defaults.ProjectType != "Node.js / Vite" || defaults.PackageManager != "pnpm" {
		t.Fatalf("project defaults = %#v", defaults)
	}
	if len(defaults.Build) != 2 || defaults.Build[0] != "pnpm" || defaults.Build[1] != "build-test" {
		t.Fatalf("build defaults = %#v", defaults.Build)
	}
	if defaults.Artifact != "dist" {
		t.Fatalf("artifact = %q", defaults.Artifact)
	}
}

func TestDiscoverFallsBackWithoutPackageFile(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "Static Site")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	defaults := Discover(directory, "test")
	if defaults.ProjectID != "static-site" || defaults.ProjectName != "Static Site" || defaults.Artifact != "" {
		t.Fatalf("defaults = %#v", defaults)
	}
	if defaults.Build != nil {
		t.Fatalf("build = %#v, want nil", defaults.Build)
	}
}
