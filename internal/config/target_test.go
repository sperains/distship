package config

import "testing"

func TestParseTargetID(t *testing.T) {
	target, ok := ParseTargetID("ipd:test")
	if !ok || target.ProjectID != "ipd" || target.EnvironmentID != "test" || target.ID() != "ipd:test" {
		t.Fatalf("ParseTargetID() = %#v, %v", target, ok)
	}
	for _, invalid := range []string{"ipd", "ipd:", ":test", "ipd:test:extra", "IPD:test"} {
		if _, ok := ParseTargetID(invalid); ok {
			t.Fatalf("ParseTargetID(%q) accepted invalid ID", invalid)
		}
	}
}

func TestFindTargetByDirectory(t *testing.T) {
	cfg := &Config{Projects: map[string]Project{
		"site": {Environments: map[string]Environment{
			"test":       {Directory: "/workspace/site"},
			"production": {Directory: "/workspace/site"},
		}},
	}}
	if !cfg.HasTargetDirectory("/workspace/site/.") {
		t.Fatal("configured directory was not found")
	}
	if _, ok := cfg.FindTargetByDirectory("/workspace/site"); ok {
		t.Fatal("ambiguous directory unexpectedly selected a target")
	}
	target, ok := cfg.FindTargetByDirectory("/workspace/site", "test")
	if !ok || target.ID() != "site:test" {
		t.Fatalf("preferred target = %#v, %v", target, ok)
	}
}

func TestIsValidSSHTarget(t *testing.T) {
	for _, valid := range []string{"bt_250", "example.com", "119.36.78.123", "root@example.com", "root@[::1]"} {
		if !IsValidSSHTarget(valid) {
			t.Errorf("IsValidSSHTarget(%q) = false", valid)
		}
	}
	for _, invalid := range []string{"", "host:22", "-oProxyCommand=x", "user name@host", "host/path", "@host"} {
		if IsValidSSHTarget(invalid) {
			t.Errorf("IsValidSSHTarget(%q) = true", invalid)
		}
	}
}
