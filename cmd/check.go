package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
		RunE: func(command *cobra.Command, args []string) error {
			return a.runCheck(command.Context(), args[0])
		},
	}
}

func (a *app) runCheck(parent context.Context, targetID string) error {
	ref, environment, err := a.resolveTarget(targetID)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
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

	labels := newSummaryLabels(a.translator)
	lines := labels.baseLines(a, environment, report)

	if report.Repository.IsRepository {
		lines = append(lines, "", a.translator.T(i18n.RecentCommitsTitle, len(report.Repository.Recent)), "")
		lines = appendCommitLines(lines, report.Repository.Recent, a.translator.T(i18n.NoRecentCommits))
	}

	lines = append(lines, "", labels.checkLine(a.translator.T(i18n.CheckSummary, gitCheckMark(report))))
	if report.RemoteState == preflight.RemoteCreatable {
		lines = append(lines, "  "+a.translator.T(i18n.CheckRemoteCreatable, report.RemoteParent))
	}
	lines = appendWarningLines(lines, a.translator.T(i18n.FieldWarnings), preflightWarnings(a, report))

	fmt.Fprintf(a.out, "%s\n\n%s\n\n%s\n\n%s\n", a.translator.T(i18n.CheckTitle, ref.ID()), status, strings.Join(lines, "\n"), a.translator.T(i18n.CheckReadOnlyNotice))
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
