package presentation

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/history"
	"github.com/sperains/distship/internal/i18n"
)

func RenderHistory(records []history.Record, targetID string, color bool, translator i18n.Translator) string {
	accent := lipgloss.NewStyle()
	if color {
		accent = accent.Foreground(lipgloss.Color("42")).Bold(true)
	}

	var output strings.Builder
	title := translator.T(i18n.HistoryTitle, len(records))
	if targetID != "" {
		title += " · " + targetID
	}
	fmt.Fprintf(&output, "%s\n\n%s\n", title, translator.T(i18n.HistoryNotice))
	if len(records) == 0 {
		if targetID == "" {
			fmt.Fprintf(&output, "\n%s\n", translator.T(i18n.HistoryEmpty))
		} else {
			fmt.Fprintf(&output, "\n%s\n", translator.T(i18n.HistoryEmptyTarget, targetID))
		}
		return output.String()
	}

	for index, record := range records {
		ref := config.TargetID(record.Project, record.Environment)
		deployedAt := record.DeployedAt.Local().Format("2006-01-02 15:04")
		fmt.Fprintf(&output, "\n%s %s · %s\n", accent.Render(fmt.Sprintf("[%d]", index+1)), ref, deployedAt)
		fmt.Fprintf(&output, "    %s\n", translator.T(i18n.FieldLine, translator.T(i18n.FieldSource), historySource(record, translator)))
		fmt.Fprintf(&output, "    %s\n", translator.T(i18n.FieldLine, translator.T(i18n.FieldDeployTarget), record.Host+":"+record.Directory))
		fmt.Fprintf(&output, "    %s\n", translator.T(i18n.FieldLine, translator.T(i18n.FieldArtifact), record.Artifact))
		fmt.Fprintf(&output, "    %s\n", translator.T(i18n.FieldLine, translator.T(i18n.FieldDuration), FormatDuration(record.Duration())))
	}
	return output.String()
}

func historySource(record history.Record, translator i18n.Translator) string {
	source := record.Branch
	if record.Commit != "" {
		commit := record.Commit
		if len(commit) > 10 {
			commit = commit[:10]
		}
		if source == "" {
			source = commit
		} else {
			source += " @ " + commit
		}
	}
	if source == "" {
		source = translator.T(i18n.NotDetected)
	}
	state := translator.T(i18n.HistoryClean)
	if record.Dirty {
		state = translator.T(i18n.HistoryDirty)
	}
	return source + " · " + state
}
