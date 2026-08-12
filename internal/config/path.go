package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sperains/distship/internal/i18n"
)

func FindPath(explicit string) (string, error) {
	paths, err := candidatePaths(explicit)
	if err != nil {
		return "", err
	}
	for _, path := range paths {
		info, statErr := os.Stat(path)
		if statErr == nil && !info.IsDir() {
			return path, nil
		}
		if statErr != nil && !os.IsNotExist(statErr) {
			return "", i18n.Wrap(i18n.ErrCheckConfig, statErr, path)
		}
	}
	return "", i18n.NewError(i18n.ErrConfigNotFound, strings.Join(paths, "\n  "))
}

func InitPath(explicit string) (string, error) {
	if explicit != "" {
		return absolutePath(explicit)
	}
	return DefaultPath()
}

func DefaultPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "distship", "projects.toml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", i18n.Wrap(i18n.ErrDetermineConfigDirectory, err)
	}
	return filepath.Join(home, ".config", "distship", "projects.toml"), nil
}

func candidatePaths(explicit string) ([]string, error) {
	if explicit != "" {
		path, err := absolutePath(explicit)
		if err != nil {
			return nil, err
		}
		return []string{path}, nil
	}
	if environment := os.Getenv("DISTSHIP_CONFIG"); environment != "" {
		path, err := absolutePath(environment)
		if err != nil {
			return nil, err
		}
		return []string{path}, nil
	}
	result := make([]string, 0, 2)
	executable, err := os.Executable()
	if err == nil {
		result = append(result, filepath.Join(filepath.Dir(executable), "projects.toml"))
	}
	defaultPath, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	result = append(result, defaultPath)
	return result, nil
}

func absolutePath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", i18n.Wrap(i18n.ErrResolveConfigPath, err)
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", i18n.Wrap(i18n.ErrResolveConfigPath, err)
	}
	return absolute, nil
}
