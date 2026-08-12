package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sperains/distship/internal/config"
	gitinfo "github.com/sperains/distship/internal/git"
	"github.com/sperains/distship/internal/history"
	"github.com/sperains/distship/internal/i18n"
	"github.com/sperains/distship/internal/preflight"
)

type deploymentView struct {
	commits       []gitinfo.Commit
	total         int
	usesHistory   bool
	noHistory     bool
	rangeWarning  i18n.Key
	previousDirty bool
}

func (a *app) deploymentView(ctx context.Context, path string, ref config.TargetRef, environment config.Environment, report preflight.Report) (deploymentView, error) {
	view := deploymentView{commits: report.Repository.Recent}
	record, found, err := history.Last(path, ref, environment.Target)
	if err != nil {
		return view, err
	}
	if !found || !report.Repository.IsRepository {
		view.noHistory = report.Repository.IsRepository
		return view, nil
	}
	view.previousDirty = record.Dirty
	rangeContext, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	changes, err := gitinfo.ChangesSince(rangeContext, a.runner, environment.Directory, record.Commit, report.Repository.RevisionHash, 5)
	if err != nil {
		return view, err
	}
	switch changes.Relation {
	case gitinfo.RangeAvailable:
		view.usesHistory = true
		view.total = changes.Total
		view.commits = changes.Commits
	case gitinfo.RangeMissing:
		view.rangeWarning = i18n.DeployHistoryMissing
	case gitinfo.RangeDiverged:
		view.rangeWarning = i18n.DeployHistoryDiverged
	}
	return view, nil
}

func printDeployPlan(a *app, ref config.TargetRef, environment config.Environment, report preflight.Report, view deploymentView, dryRun bool) {
	warnings := deployWarnings(a, report, view)
	status := "✓ " + a.translator.T(i18n.DeployReady)
	if dryRun {
		status = "◇ " + a.translator.T(i18n.DeployDryRun)
	} else if len(warnings) > 0 {
		status = "! " + a.translator.T(i18n.DeployReadyWithWarnings)
	}

	labels := newSummaryLabels(a.translator)
	lines := append(labels.baseLines(a, environment, report), "")
	if report.Repository.IsRepository {
		title := a.translator.T(i18n.DeployRecentCommits, len(view.commits))
		if view.usesHistory {
			title = a.translator.T(i18n.DeployChangesSince, view.total)
		}
		lines = append(lines, title, "")
		emptyMessage := a.translator.T(i18n.NoRecentCommits)
		if view.usesHistory {
			emptyMessage = a.translator.T(i18n.DeployNoNewCommits)
		}
		lines = appendCommitLines(lines, view.commits, emptyMessage)
	}

	checkSummary := a.translator.T(i18n.DeployChecks, gitCheckMark(report))
	if dryRun {
		checkSummary = a.translator.T(i18n.DeployDryRunChecks, gitCheckMark(report))
	}
	lines = append(lines, "", labels.checkLine(checkSummary))
	if view.noHistory {
		lines = append(lines, "", a.translator.T(i18n.DeployNoLocalHistory))
	}
	if !dryRun && report.RemoteState == preflight.RemoteCreatable {
		lines = append(lines, a.translator.T(i18n.CheckRemoteCreatable, report.RemoteParent))
	}
	lines = appendWarningLines(lines, a.translator.T(i18n.FieldWarnings), warnings)

	notice := a.translator.T(i18n.DeployActionNotice)
	if dryRun {
		notice = a.translator.T(i18n.DeployDryRunNotice)
	}
	fmt.Fprintf(a.out, "%s\n\n%s\n\n%s\n\n%s\n", a.translator.T(i18n.DeployTitle, ref.ID()), status, strings.Join(lines, "\n"), notice)
}

func deployWarnings(a *app, report preflight.Report, view deploymentView) []string {
	warnings := preflightWarnings(a, report)
	if view.rangeWarning != "" {
		warnings = append(warnings, a.translator.T(view.rangeWarning))
	}
	if view.previousDirty {
		warnings = append(warnings, a.translator.T(i18n.DeployPreviousDirty))
	}
	return warnings
}
