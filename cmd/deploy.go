package cmd

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	buildtool "github.com/sperains/distship/internal/build"
	"github.com/sperains/distship/internal/deployment"
	"github.com/sperains/distship/internal/history"
	"github.com/sperains/distship/internal/i18n"
	"github.com/sperains/distship/internal/preflight"
	"github.com/sperains/distship/internal/presentation"
	"github.com/sperains/distship/internal/transport"
)

func (a *app) newDeployCommand() *cobra.Command {
	var dryRun bool
	var yes bool
	command := &cobra.Command{
		Use:     "deploy <target-id>",
		Short:   a.translator.T(i18n.DeployShort),
		Example: "  distship deploy ipd:test",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return i18n.NewError(i18n.InvalidTargetID)
			}
			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			return a.runDeploy(command.Context(), args[0], dryRun, yes)
		},
	}
	command.Flags().BoolVar(&dryRun, "dry-run", false, a.translator.T(i18n.FlagDryRun))
	command.Flags().BoolVar(&yes, "yes", false, a.translator.T(i18n.FlagYes))
	return command
}

func (a *app) runDeploy(ctx context.Context, targetID string, dryRun, yes bool) error {
	ref, environment, err := a.resolveTarget(targetID)
	if err != nil {
		return err
	}

	checkContext, cancelCheck := context.WithTimeout(ctx, 20*time.Second)
	var report preflight.Report
	if dryRun {
		report, err = preflight.CheckLocal(checkContext, a.runner, environment)
	} else {
		report, err = preflight.Check(checkContext, a.runner, environment)
	}
	cancelCheck()
	if err != nil {
		return a.localizeCheckError(err)
	}
	if !dryRun {
		if err := transport.CheckAvailable(a.runner); err != nil {
			return i18n.NewError(i18n.DeployRsyncUnavailable, err)
		}
	}

	historyPath, err := history.DefaultPath()
	if err != nil {
		return i18n.NewError(i18n.DeployHistoryFailed, err)
	}
	view, err := a.deploymentView(ctx, historyPath, ref, environment, report)
	if err != nil {
		return i18n.NewError(i18n.DeployHistoryFailed, err)
	}
	printDeployPlan(a, ref, environment, report, view, dryRun)
	if dryRun {
		return nil
	}

	if !yes {
		answer, err := prompt(bufio.NewReader(a.in), a.out, a.translator.T(i18n.ConfirmDeploy), "N")
		if err != nil {
			return err
		}
		if !isYes(answer) {
			fmt.Fprintln(a.out, a.translator.T(i18n.DeployCancelled))
			return nil
		}
	}

	lock, err := deployment.Acquire(filepath.Dir(historyPath), ref.ID())
	if err != nil {
		return i18n.NewError(i18n.DeployLockFailed, err)
	}
	lockReleased := false
	defer func() {
		if !lockReleased {
			_ = lock.Release()
		}
	}()

	started := time.Now()
	fmt.Fprintf(a.out, "\n%s\n\n", a.translator.T(i18n.BuildTitle))
	buildResult, err := buildtool.Run(ctx, a.runner, environment, a.out)
	if err != nil {
		return i18n.NewError(i18n.DeployBuildFailed, err)
	}
	fmt.Fprintf(a.out, "\n✓ %s · %s\n", a.translator.T(i18n.BuildCompleted), presentation.FormatDuration(buildResult.Duration))

	if report.RemoteState == preflight.RemoteCreatable {
		if err := transport.EnsureRemote(ctx, a.runner, environment.Target); err != nil {
			return i18n.NewError(i18n.DeployRemoteCreateFailed, err)
		}
	}
	fmt.Fprintf(a.out, "\n%s\n\n", a.translator.T(i18n.UploadTitle))
	if err := transport.Upload(ctx, a.runner, buildResult.Artifact, environment.Target, a.out); err != nil {
		return i18n.NewError(i18n.DeployUploadFailed, err)
	}

	duration := time.Since(started)
	record := history.Record{
		DeployedAt:    time.Now(),
		Project:       ref.ProjectID,
		Environment:   ref.EnvironmentID,
		Host:          environment.Target.Host,
		Directory:     environment.Target.Directory,
		Branch:        report.Repository.Branch,
		Commit:        report.Repository.RevisionHash,
		Dirty:         report.Repository.Changes.Total() > 0,
		Artifact:      environment.Artifact,
		DurationMilli: duration.Milliseconds(),
	}
	if err := history.Append(historyPath, record); err != nil {
		return i18n.NewError(i18n.DeployHistoryWriteFailed, err)
	}
	if err := lock.Release(); err != nil {
		return i18n.NewError(i18n.DeployLockReleaseFailed, err)
	}
	lockReleased = true

	fmt.Fprintf(a.out, "\n✓ %s\n\n  %s\n  %s\n  %s\n", a.translator.T(i18n.DeployCompleted), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldDeployTarget), environment.Target.String()), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldDuration), presentation.FormatDuration(duration)), "✓ "+a.translator.T(i18n.DeployHistoryWritten))
	return nil
}
