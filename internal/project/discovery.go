package project

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Defaults struct {
	ProjectID      string
	ProjectName    string
	ProjectType    string
	PackageManager string
	Build          []string
	Artifact       string
	Branch         string
}

type packageFile struct {
	Name           string            `json:"name"`
	PackageManager string            `json:"packageManager"`
	Scripts        map[string]string `json:"scripts"`
}

func Discover(directory, environment string) Defaults {
	base := filepath.Base(filepath.Clean(directory))
	result := Defaults{
		ProjectID:   identifier(base),
		ProjectName: base,
		Artifact:    detectExistingArtifact(directory),
		Branch:      detectBranch(directory),
	}
	packageData, err := os.ReadFile(filepath.Join(directory, "package.json"))
	if err != nil {
		return result
	}
	var manifest packageFile
	if json.Unmarshal(packageData, &manifest) != nil {
		return result
	}
	result.ProjectType = "Node.js"
	if strings.TrimSpace(manifest.Name) != "" {
		result.ProjectName = manifest.Name
		result.ProjectID = identifier(manifest.Name)
	}
	manager := detectPackageManager(directory, manifest.PackageManager)
	result.PackageManager = manager
	if script := detectBuildScript(manifest.Scripts, environment); script != "" {
		result.Build = buildCommand(manager, script)
		command := manifest.Scripts[script]
		if strings.Contains(command, "vite build") {
			result.ProjectType = "Node.js / Vite"
			if result.Artifact == "" {
				result.Artifact = "dist"
			}
		}
		if strings.Contains(command, "react-scripts build") && result.Artifact == "" {
			result.Artifact = "build"
		}
	}
	return result
}

func identifier(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var result strings.Builder
	separator := false
	for _, character := range value {
		switch {
		case character >= 'a' && character <= 'z' || character >= '0' && character <= '9':
			if separator && result.Len() > 0 {
				result.WriteByte('-')
			}
			result.WriteRune(character)
			separator = false
		default:
			separator = true
		}
	}
	return strings.Trim(result.String(), "-")
}

func detectPackageManager(directory, declared string) string {
	if manager := strings.TrimSpace(strings.SplitN(declared, "@", 2)[0]); manager != "" {
		return manager
	}
	for _, candidate := range []struct{ file, manager string }{
		{"pnpm-lock.yaml", "pnpm"},
		{"bun.lock", "bun"},
		{"bun.lockb", "bun"},
		{"yarn.lock", "yarn"},
		{"package-lock.json", "npm"},
	} {
		if _, err := os.Stat(filepath.Join(directory, candidate.file)); err == nil {
			return candidate.manager
		}
	}
	return "npm"
}

func detectBuildScript(scripts map[string]string, environment string) string {
	for _, candidate := range []string{"build-" + environment, "build:" + environment, "build_" + environment, "build"} {
		if _, exists := scripts[candidate]; exists {
			return candidate
		}
	}
	var buildScripts []string
	for name := range scripts {
		if strings.HasPrefix(name, "build") {
			buildScripts = append(buildScripts, name)
		}
	}
	if len(buildScripts) == 1 {
		return buildScripts[0]
	}
	sort.Strings(buildScripts)
	return ""
}

func buildCommand(manager, script string) []string {
	switch manager {
	case "pnpm", "yarn":
		return []string{manager, script}
	case "bun":
		return []string{"bun", "run", script}
	default:
		return []string{"npm", "run", script}
	}
}

func detectExistingArtifact(directory string) string {
	for _, candidate := range []string{"dist", "build", "out"} {
		if info, err := os.Stat(filepath.Join(directory, candidate)); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

func detectBranch(directory string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "git", "-C", directory, "branch", "--show-current").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
