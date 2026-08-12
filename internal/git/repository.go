package git

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sperains/distship/internal/process"
)

type Commit struct {
	Hash      string
	ShortHash string
	Subject   string
	Author    string
	Time      time.Time
}

type Changes struct {
	Modified  int
	Added     int
	Deleted   int
	Untracked int
}

func (c Changes) Total() int {
	return c.Modified + c.Added + c.Deleted + c.Untracked
}

type Repository struct {
	IsRepository bool
	Branch       string
	Detached     bool
	RevisionHash string
	Revision     string
	Recent       []Commit
	Changes      Changes
}

func Inspect(ctx context.Context, runner process.Runner, directory string) (Repository, error) {
	inside, err := run(ctx, runner, directory, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		if strings.Contains(inside.Stderr, "not a git repository") {
			return Repository{}, nil
		}
		return Repository{}, commandError("detect repository", inside, err)
	}
	if strings.TrimSpace(inside.Stdout) != "true" {
		return Repository{}, nil
	}

	repository := Repository{IsRepository: true}
	branch, branchErr := run(ctx, runner, directory, "symbolic-ref", "--quiet", "--short", "HEAD")
	if branchErr != nil {
		repository.Detached = true
	} else {
		repository.Branch = strings.TrimSpace(branch.Stdout)
	}

	revision, err := run(ctx, runner, directory, "log", "-1", "--format=%H%x00%h", "HEAD")
	if err != nil {
		return Repository{}, commandError("read current revision", revision, err)
	}
	revisionParts := strings.Split(strings.TrimSpace(revision.Stdout), "\x00")
	if len(revisionParts) != 2 {
		return Repository{}, fmt.Errorf("read current revision: unexpected git output")
	}
	repository.RevisionHash = revisionParts[0]
	repository.Revision = revisionParts[1]

	recent, err := run(ctx, runner, directory, "log", "--no-merges", "-3", "--format=%H%x1f%h%x1f%s%x1f%an%x1f%aI%x1e", "HEAD")
	if err != nil {
		return Repository{}, commandError("read recent changes", recent, err)
	}
	repository.Recent, err = parseCommits(recent.Stdout)
	if err != nil {
		return Repository{}, err
	}

	status, err := run(ctx, runner, directory, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return Repository{}, commandError("read working tree", status, err)
	}
	repository.Changes = parseStatus(status.Stdout)
	return repository, nil
}

func parseCommits(output string) ([]Commit, error) {
	records := strings.Split(output, "\x1e")
	commits := make([]Commit, 0, len(records))
	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}
		parts := strings.Split(record, "\x1f")
		if len(parts) != 5 {
			return nil, fmt.Errorf("read recent changes: unexpected git output")
		}
		committedAt, err := time.Parse(time.RFC3339, parts[4])
		if err != nil {
			return nil, fmt.Errorf("read recent change time: %w", err)
		}
		commits = append(commits, Commit{Hash: parts[0], ShortHash: parts[1], Subject: parts[2], Author: parts[3], Time: committedAt})
	}
	return commits, nil
}

func run(ctx context.Context, runner process.Runner, directory string, args ...string) (process.Result, error) {
	return runner.Run(ctx, process.Command{Name: "git", Args: append([]string{"-C", directory}, args...)})
}

func commandError(action string, result process.Result, err error) error {
	detail := strings.TrimSpace(result.Stderr)
	if detail == "" {
		detail = err.Error()
	}
	return fmt.Errorf("%s: %s", action, detail)
}

func parseStatus(output string) Changes {
	var changes Changes
	entries := strings.Split(output, "\x00")
	for index := 0; index < len(entries); index++ {
		entry := entries[index]
		if len(entry) < 3 {
			continue
		}
		status := entry[:2]
		switch {
		case status == "??":
			changes.Untracked++
		case strings.Contains(status, "D"):
			changes.Deleted++
		case strings.Contains(status, "A"):
			changes.Added++
		default:
			changes.Modified++
		}
		if strings.ContainsAny(status, "RC") {
			index++
		}
	}
	return changes
}
