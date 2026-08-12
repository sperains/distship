package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/sperains/distship/internal/config"
	gitinfo "github.com/sperains/distship/internal/git"
	"github.com/sperains/distship/internal/i18n"
	"github.com/sperains/distship/internal/preflight"
)

type summaryLabels struct {
	source string
	target string
	build  string
	checks string
	width  int
}

func newSummaryLabels(translator i18n.Translator) summaryLabels {
	labels := summaryLabels{
		source: translator.T(i18n.FieldSource),
		target: translator.T(i18n.FieldDeployTarget),
		build:  translator.T(i18n.FieldBuild),
		checks: translator.T(i18n.FieldChecks),
	}
	for _, label := range []string{labels.source, labels.target, labels.build, labels.checks} {
		labels.width = max(labels.width, lipgloss.Width(label))
	}
	return labels
}

func (l summaryLabels) baseLines(a *app, environment config.Environment, report preflight.Report) []string {
	return []string{
		summaryLine(l.source, l.width, sourceSummary(a, environment, report)),
		summaryLine(l.target, l.width, environment.Target.String()),
		summaryLine(l.build, l.width, strings.Join(environment.Build, " ")+" → "+environment.Artifact),
	}
}

func (l summaryLabels) checkLine(value string) string {
	return summaryLine(l.checks, l.width, value)
}

func sourceSummary(a *app, environment config.Environment, report preflight.Report) string {
	if !report.Repository.IsRepository {
		return environment.Directory
	}
	branch := report.Repository.Branch
	if report.Repository.Detached {
		branch = a.translator.T(i18n.DetachedHead)
	}
	workingTree := a.translator.T(i18n.WorkingTreeClean)
	if total := report.Repository.Changes.Total(); total > 0 {
		workingTree = a.translator.T(i18n.WorkingTreeDirty, total)
	}
	return fmt.Sprintf("%s @ %s · %s", branch, report.Repository.Revision, workingTree)
}

func gitCheckMark(report preflight.Report) string {
	if report.Repository.IsRepository {
		return "✓"
	}
	return "!"
}

func appendCommitLines(lines []string, commits []gitinfo.Commit, emptyMessage string) []string {
	if len(commits) == 0 {
		return append(lines, "  "+emptyMessage)
	}
	for index, commit := range commits {
		lines = append(lines,
			"  "+commit.Subject,
			fmt.Sprintf("  %s · %s · %s", commit.ShortHash, commit.Author, commit.Time.Format("01-02 15:04")),
		)
		if index < len(commits)-1 {
			lines = append(lines, "")
		}
	}
	return lines
}

func preflightWarnings(a *app, report preflight.Report) []string {
	var warnings []string
	if report.HasWarning(preflight.WarningGitUnavailable) {
		warnings = append(warnings, a.translator.T(i18n.WarningGitUnavailable))
	}
	if report.HasWarning(preflight.WarningNotGit) {
		warnings = append(warnings, a.translator.T(i18n.WarningNotGit))
	}
	if report.HasWarning(preflight.WarningDirty) {
		warnings = append(warnings, a.translator.T(i18n.WarningDirtyWorkingTree))
	}
	if report.HasWarning(preflight.WarningDetached) {
		warnings = append(warnings, a.translator.T(i18n.WarningDetachedHead))
	}
	return warnings
}

func appendWarningLines(lines []string, title string, warnings []string) []string {
	if len(warnings) == 0 {
		return lines
	}
	lines = append(lines, "", title)
	for _, warning := range warnings {
		lines = append(lines, "  ! "+warning)
	}
	return lines
}

func summaryLine(label string, width int, value string) string {
	return label + strings.Repeat(" ", width-lipgloss.Width(label)+2) + value
}
