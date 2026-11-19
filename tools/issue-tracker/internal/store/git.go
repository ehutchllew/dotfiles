package store

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/doc"
)

// ResolveCommit turns a ref such as HEAD into a full commit SHA. It only
// reads git. A full SHA is returned as given without calling git.
func (l *Local) ResolveCommit(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if doc.IsFullSHA(ref) {
		return ref, nil
	}
	if ref == "" || strings.HasPrefix(ref, "-") {
		return "", apperr.Usage("invalid_commit", "%q is not a commit", ref)
	}
	out, err := l.git("rev-parse", "--verify", "--quiet", "--end-of-options", ref+"^{commit}")
	if err != nil || !doc.IsFullSHA(out) {
		return "", apperr.New("invalid_commit", "git can't resolve %q to a commit", ref)
	}
	return out, nil
}

// CheckHandoffCheckout checks that the stopped attempt's source is still
// checked out. Committed tracker changes are allowed after the candidate.
func (l *Local) CheckHandoffCheckout(commit string) error {
	status, err := l.git("status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return apperr.New("handoff_checkout", "git couldn't check the working tree: %v", err)
	}
	if status != "" {
		return apperr.New("handoff_checkout", "pending hand-off needs a clean checkout")
	}
	if _, err := l.git("merge-base", "--is-ancestor", commit, "HEAD"); err != nil {
		return apperr.New("stale_handoff", "hand-off commit %s is not in HEAD's history", commit)
	}
	changed, err := l.git("diff", "--name-only", commit, "HEAD", "--", ".",
		":(exclude,glob,top).scratch/*/spec.md", ":(exclude,glob,top).scratch/*/issues/*.md")
	if err != nil {
		return apperr.New("handoff_checkout", "git couldn't compare the hand-off commit to HEAD: %v", err)
	}
	if changed != "" {
		return apperr.New("stale_handoff", "files other than specs and tickets differ from hand-off commit %s", commit).With("changed", strings.Split(changed, "\n"))
	}
	return nil
}

// IsIgnored reports whether git ignores a repo-relative path. ok is false
// when git can't answer.
func (l *Local) IsIgnored(rel string) (ignored, ok bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", l.Root, "check-ignore", "-q", "--", rel)
	cmd.Env = append(cmd.Environ(), "GIT_OPTIONAL_LOCKS=0")
	err := cmd.Run()
	if err == nil {
		return true, true
	}
	if ee, isExit := err.(*exec.ExitError); isExit && ee.ExitCode() == 1 {
		return false, true
	}
	return false, false
}

func (l *Local) git(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", l.Root}, args...)...)
	cmd.Env = append(cmd.Environ(), "GIT_OPTIONAL_LOCKS=0")
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}
