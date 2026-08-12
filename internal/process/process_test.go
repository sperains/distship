package process

import (
	"context"
	"strings"
	"testing"
)

func TestOSRunnerCapturesOutputAndUsesDirectory(t *testing.T) {
	directory := t.TempDir()
	result, err := (OSRunner{}).Run(context.Background(), Command{
		Name:  "sh",
		Args:  []string{"-c", `printf '%s:%s' "$PWD" "$(cat)"`},
		Dir:   directory,
		Stdin: "input",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(result.Stdout, directory+":input") {
		t.Fatalf("stdout = %q", result.Stdout)
	}
}
