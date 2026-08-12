package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/i18n"
	"github.com/sperains/distship/internal/preflight"
)

func (a *app) newCheckCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "check <target-id>",
		Short:   a.translator.T(i18n.CheckShort),
		Example: "  distship check ipd:test",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return i18n.NewError(i18n.InvalidTargetID)
			}
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			return a.runCheck(args[0])
		},
	}
}

func (a *app) runCheck(targetID string) error {
	ref, ok := config.ParseTargetID(targetID)
	if !ok {
		return i18n.NewError(i18n.InvalidTargetID)
	}
	cfg, _, err := a.loadConfig()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return i18n.NewError(i18n.ConfigInvalid, a.translator.Error(err))
	}
	_, environment, exists := cfg.Target(ref)
	if !exists {
		return i18n.NewError(i18n.TargetNotFound, ref.ID(), strings.Join(cfg.TargetIDs(), "\n  "))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	report, err := preflight.Check(ctx, a.runner, environment)
	if err != nil {
		return a.localizeCheckError(err)
	}
	printCheckReport(a, ref, environment, report)
	return nil
}

func printCheckReport(a *app, ref config.TargetRef, environment config.Environment, report preflight.Report) {
	status := "✓ " + a.translator.T(i18n.CheckPassed)
	if len(report.Warnings) > 0 {
		status = "! " + a.translator.T(i18n.CheckPassedWithWarnings)
	}

	workingTree := a.translator.T(i18n.WorkingTreeClean)
	if total := report.Repository.Changes.Total(); total > 0 {
		workingTree = a.translator.T(i18n.WorkingTreeDirty, total)
	}
	source := environment.Directory
	if report.Repository.IsRepository {
		branch := report.Repository.Branch
		if report.Repository.Detached {
			branch = a.translator.T(i18n.DetachedHead)
		}
		source = fmt.Sprintf("%s @ %s · %s", branch, report.Repository.Revision, workingTree)
	} else if containsString(report.Warnings, "git_unavailable") {
		source += " · " + a.translator.T(i18n.WarningGitUnavailable)
	} else {
		source += " · " + a.translator.T(i18n.WarningNotGit)
	}

	labels := []string{
		a.translator.T(i18n.FieldSource),
		a.translator.T(i18n.FieldDeployTarget),
		a.translator.T(i18n.FieldBuild),
		a.translator.T(i18n.FieldChecks),
	}
	labelWidth := widestLabel(labels)
	lines := []string{
		summaryLine(labels[0], labelWidth, source),
		summaryLine(labels[1], labelWidth, environment.Target.String()),
		summaryLine(labels[2], labelWidth, strings.Join(environment.Build, " ")+" → "+environment.Artifact),
	}

	if report.Repository.IsRepository {
		lines = append(lines, "", a.translator.T(i18n.RecentCommitsTitle, len(report.Repository.Recent)), "")
		if len(report.Repository.Recent) == 0 {
			lines = append(lines, "  "+a.translator.T(i18n.NoRecentCommits))
		} else {
			for index, commit := range report.Repository.Recent {
				lines = append(lines,
					"  "+commit.Subject,
					fmt.Sprintf("  %s · %s · %s", commit.ShortHash, commit.Author, commit.Time.Format("01-02 15:04")),
				)
				if index < len(report.Repository.Recent)-1 {
					lines = append(lines, "")
				}
			}
		}
	}

	gitCheck := "✓"
	if !report.Repository.IsRepository {
		gitCheck = "!"
	}
	lines = append(lines, "", summaryLine(labels[3], labelWidth, a.translator.T(i18n.CheckSummary, gitCheck)))
	if report.RemoteState == preflight.RemoteExisting {
		// The compact check summary is enough for the normal path.
	} else if report.RemoteState == preflight.RemoteCreatable {
		lines = append(lines, "  "+a.translator.T(i18n.CheckRemoteCreatable, report.RemoteParent))
	}

	var warnings []string
	if containsString(report.Warnings, "git_unavailable") {
		warnings = append(warnings, a.translator.T(i18n.WarningGitUnavailable))
	}
	if containsString(report.Warnings, "not_git") {
		warnings = append(warnings, a.translator.T(i18n.WarningNotGit))
	}
	if containsString(report.Warnings, "dirty") {
		warnings = append(warnings, a.translator.T(i18n.WarningDirtyWorkingTree))
	}
	if containsString(report.Warnings, "detached") {
		warnings = append(warnings, a.translator.T(i18n.WarningDetachedHead))
	}
	if len(warnings) > 0 {
		lines = append(lines, "", a.translator.T(i18n.FieldWarnings))
		for _, warning := range warnings {
			lines = append(lines, "  ! "+warning)
		}
	}

	fmt.Fprintf(a.out, "%s\n\n%s\n\n%s\n\n%s\n", a.translator.T(i18n.CheckTitle, ref.ID()), status, strings.Join(lines, "\n"), a.translator.T(i18n.CheckReadOnlyNotice))
}

func widestLabel(labels []string) int {
	width := 0
	for _, label := range labels {
		width = max(width, lipgloss.Width(label))
	}
	return width
}

func summaryLine(label string, width int, value string) string {
	return label + strings.Repeat(" ", width-lipgloss.Width(label)+2) + value
}

func (a *app) localizeCheckError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return i18n.NewError(i18n.CheckTimedOut)
	}
	var checkError *preflight.CheckError
	if !errors.As(err, &checkError) {
		return err
	}
	keys := map[preflight.ErrorKind]i18n.Key{
		preflight.ErrorLocalDirectory: i18n.CheckLocalDirectoryFailed,
		preflight.ErrorBuildTool:      i18n.CheckBuildToolFailed,
		preflight.ErrorGit:            i18n.CheckGitFailed,
		preflight.ErrorBranch:         i18n.CheckBranchFailed,
		preflight.ErrorDirty:          i18n.CheckDirtyDenied,
		preflight.ErrorSSH:            i18n.CheckSSHFailed,
		preflight.ErrorRemote:         i18n.CheckRemoteFailed,
	}
	return i18n.NewError(keys[checkError.Kind], checkError.Detail)
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
