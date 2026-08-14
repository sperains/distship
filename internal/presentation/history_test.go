package presentation

import (
	"strings"
	"testing"
	"time"

	"github.com/sperains/distship/internal/history"
	"github.com/sperains/distship/internal/i18n"
)

func TestRenderHistoryUsesCompactCards(t *testing.T) {
	records := []history.Record{{
		DeployedAt:    time.Date(2026, 8, 12, 10, 4, 0, 0, time.Local),
		Project:       "site",
		Environment:   "test",
		Host:          "staging-web",
		Directory:     "/var/www/site",
		Branch:        "main",
		Commit:        "a6f0684b12345678",
		Artifact:      "dist",
		DurationMilli: 12400,
	}}
	output := RenderHistory(records, "site:test", false, i18n.New(i18n.English))
	for _, expected := range []string{
		"Local deployment history · 1 shown · site:test",
		"Local records only; they may not match the server's current state.",
		"[1] site:test · 2026-08-12 10:04",
		"Source: main @ a6f0684b12 · clean",
		"Target: staging-web:/var/www/site",
		"Artifact: dist",
		"Duration: 12.4s",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("RenderHistory() missing %q:\n%s", expected, output)
		}
	}
}

func TestRenderHistoryShowsLocalizedEmptyState(t *testing.T) {
	output := RenderHistory(nil, "site:test", false, i18n.New(i18n.SimplifiedChinese))
	for _, expected := range []string{"本机部署历史 · 展示 0 条 · site:test", "仅包含本机记录", "暂无 site:test 的本机部署历史。"} {
		if !strings.Contains(output, expected) {
			t.Errorf("RenderHistory() missing %q:\n%s", expected, output)
		}
	}
}

func TestRenderHistoryMarksDirtyDeployments(t *testing.T) {
	output := RenderHistory([]history.Record{{Project: "site", Environment: "test", Dirty: true}}, "", false, i18n.New(i18n.SimplifiedChinese))
	if !strings.Contains(output, "来源：未识别 · 包含未提交改动") {
		t.Fatalf("RenderHistory() = %s", output)
	}
}
