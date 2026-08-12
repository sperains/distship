package presentation

import (
	"strings"
	"testing"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/i18n"
)

func TestRenderProjectsUsesGroupedCards(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Projects: map[string]config.Project{
			"ipd": {
				Name: "IPD",
				Environments: map[string]config.Environment{
					"test": {
						Name:      "测试环境",
						Directory: "/Users/example/ipd",
						Target:    config.Target{Host: "staging-web", Directory: "/var/www/ipd"},
						Git:       config.GitPolicy{AllowedBranches: []string{"test"}},
					},
				},
			},
		},
	}

	output := RenderProjects(cfg, false, i18n.New(i18n.SimplifiedChinese))
	for _, expected := range []string{"[1] IPD · 测试环境", "标识：ipd:test", "本地：/Users/example/ipd", "远端：staging-web:/var/www/ipd", "分支：test"} {
		if !strings.Contains(output, expected) {
			t.Errorf("RenderProjects() missing %q:\n%s", expected, output)
		}
	}
	if strings.Contains(output, "\x1b[") {
		t.Fatalf("RenderProjects() contains ANSI escapes when color is disabled: %q", output)
	}
}

func TestRenderProjectsInEnglish(t *testing.T) {
	cfg := &config.Config{
		Version: 1,
		Projects: map[string]config.Project{
			"site": {
				Name: "Website",
				Environments: map[string]config.Environment{
					"staging": {
						Name:      "Staging",
						Directory: "/srv/site",
						Target:    config.Target{Host: "web", Directory: "/var/www/site"},
					},
				},
			},
		},
	}
	output := RenderProjects(cfg, false, i18n.New(i18n.English))
	for _, expected := range []string{"[1] Website · Staging", "ID: site:staging", "Local: /srv/site", "Remote: web:/var/www/site", "Branches: any branch"} {
		if !strings.Contains(output, expected) {
			t.Errorf("RenderProjects() missing %q:\n%s", expected, output)
		}
	}
}
