package i18n

import (
	"errors"
	"strings"
	"testing"
)

func TestResolvePriorityAndNormalization(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "")
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "zh_CN.UTF-8")
	translator, err := Resolve("")
	if err != nil {
		t.Fatal(err)
	}
	if translator.Language() != SimplifiedChinese {
		t.Fatalf("language = %s, want zh-CN", translator.Language())
	}

	t.Setenv("DISTSHIP_LANG", "zh-CN")
	translator, err = Resolve("en-US")
	if err != nil {
		t.Fatal(err)
	}
	if translator.Language() != English {
		t.Fatalf("explicit language = %s, want en", translator.Language())
	}
}

func TestResolveFallsBackToEnglish(t *testing.T) {
	t.Setenv("DISTSHIP_LANG", "")
	t.Setenv("LC_ALL", "fr_FR.UTF-8")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "")
	translator, err := Resolve("")
	if err != nil {
		t.Fatal(err)
	}
	if translator.Language() != English {
		t.Fatalf("language = %s, want en", translator.Language())
	}
}

func TestLocalizedWrappedError(t *testing.T) {
	err := Wrap(ErrReadConfig, errors.New("permission denied"), "/tmp/projects.toml")
	message := New(SimplifiedChinese).Error(err)
	if !strings.Contains(message, "读取配置 /tmp/projects.toml") || !strings.Contains(message, "permission denied") {
		t.Fatalf("localized error = %q", message)
	}
}

func TestUnsupportedExplicitLanguage(t *testing.T) {
	if _, err := Resolve("fr"); err == nil {
		t.Fatal("Resolve(fr) error = nil")
	}
}

func TestCatalogsContainEveryEnglishMessage(t *testing.T) {
	for language, catalog := range catalogs {
		for key := range catalogs[English] {
			if _, exists := catalog[key]; !exists {
				t.Errorf("catalog %s is missing %s", language, key)
			}
		}
	}
}
