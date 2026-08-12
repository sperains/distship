package git

import "testing"

func TestParseStatus(t *testing.T) {
	changes := parseStatus(" M changed.go\x00A  added.go\x00 D deleted.go\x00?? new.go\x00R  renamed.go\x00old.go\x00")
	if changes.Modified != 2 || changes.Added != 1 || changes.Deleted != 1 || changes.Untracked != 1 || changes.Total() != 5 {
		t.Fatalf("changes = %#v", changes)
	}
}

func TestParseCommits(t *testing.T) {
	output := "aaa\x1fa1\x1ffeat: first\x1fAlice\x1f2026-08-12T10:30:00+08:00\x1e\n" +
		"bbb\x1fb2\x1ffix: second\x1fBob\x1f2026-08-11T09:20:00+08:00\x1e\n"
	commits, err := parseCommits(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 || commits[0].Subject != "feat: first" || commits[1].Author != "Bob" {
		t.Fatalf("commits = %#v", commits)
	}
	if commits[0].Time.Format("2006-01-02 15:04") != "2026-08-12 10:30" {
		t.Fatalf("commit time = %s", commits[0].Time)
	}
}

func TestParseCommitsAllowsNoNonMergeHistory(t *testing.T) {
	commits, err := parseCommits("")
	if err != nil || len(commits) != 0 {
		t.Fatalf("commits = %#v, error = %v", commits, err)
	}
}
