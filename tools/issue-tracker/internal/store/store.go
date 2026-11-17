// Package store finds the target repository, selects the backend, and reads
// and writes feature documents in local storage.
package store

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/doc"
)

// ConfigPath is the repo-relative backend selection file.
const ConfigPath = "docs/agents/issue-tracker.md"

// ScratchDir is the repo-relative root of local storage.
const ScratchDir = ".scratch"

// FindRepo walks up from start to the nearest directory with a `.git` entry.
// A worktree or submodule has a `.git` file, so it resolves to its own root.
func FindRepo(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", apperr.New("repo_not_found", "can't resolve %s: %v", start, err)
	}
	for {
		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", apperr.New("repo_not_found", "no git working tree contains %s; run inside a repository or pass --repo", start)
		}
		dir = parent
	}
}

// Backend reads the configured backend name, defaulting to "local".
func Backend(root string) (string, error) {
	f, err := os.Open(filepath.Join(root, ConfigPath))
	if errors.Is(err, fs.ErrNotExist) {
		return "local", nil
	}
	if err != nil {
		return "", apperr.New("invalid_config", "can't read %s: %v", ConfigPath, err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if v, ok := strings.CutPrefix(line, "Backend:"); ok {
			v = strings.ToLower(strings.Trim(strings.TrimSpace(v), "`"))
			if v == "" {
				return "", apperr.New("invalid_config", "%s has an empty `Backend:` line", ConfigPath)
			}
			return v, nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", apperr.New("invalid_config", "can't read %s: %v", ConfigPath, err)
	}
	return "local", nil
}

// Local stores features under `.scratch/` in one working tree.
type Local struct {
	Root      string
	Canonical string
	LockDir   string
}

// Open selects the backend for root. Only the local backend exists.
func Open(root string) (*Local, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, apperr.New("repo_not_found", "can't resolve %s: %v", root, err)
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return nil, apperr.New("repo_not_found", "%s is not a directory", root)
	}
	backend, err := Backend(abs)
	if err != nil {
		return nil, err
	}
	if backend != "local" {
		return nil, apperr.New("unsupported_backend", "backend %q from %s is not supported; only `local` is", backend, ConfigPath)
	}
	canonical, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return nil, apperr.New("repo_not_found", "can't resolve %s: %v", root, err)
	}
	return &Local{Root: abs, Canonical: canonical, LockDir: defaultLockDir()}, nil
}

var featureIDRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)

// CheckFeatureID rejects IDs that aren't a single safe path segment.
func CheckFeatureID(id string) error {
	if !featureIDRe.MatchString(id) || strings.Contains(id, "..") {
		return apperr.Usage("invalid_feature_id", "feature ID %q must be a slug such as `order-summary`", id)
	}
	return nil
}

// FeatureDir returns the repo-relative feature directory.
func FeatureDir(id string) string { return filepath.ToSlash(filepath.Join(ScratchDir, id)) }

// SpecPath returns the repo-relative spec path.
func SpecPath(id string) string { return FeatureDir(id) + "/spec.md" }

// IssuesDir returns the repo-relative ticket directory.
func IssuesDir(id string) string { return FeatureDir(id) + "/issues" }

// ListFeatures returns the IDs of directories in `.scratch/` with a spec.
func (l *Local) ListFeatures() ([]string, error) {
	if err := l.checkPath(ScratchDir); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(l.Root, ScratchDir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, apperr.New("read_failed", "can't list %s: %v", ScratchDir, err)
	}
	var ids []string
	for _, e := range entries {
		if !e.IsDir() || CheckFeatureID(e.Name()) != nil {
			continue
		}
		if info, err := os.Lstat(filepath.Join(l.Root, SpecPath(e.Name()))); err == nil && info.Mode().IsRegular() {
			ids = append(ids, e.Name())
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// TicketFile is one ticket file name in a feature's issues directory.
type TicketFile struct {
	Number int
	Path   string
}

var ticketFileRe = regexp.MustCompile(`^(\d+)-[^/]+\.md$`)

// ListTicketFiles returns ticket files sorted by number. Markdown files that
// don't follow the naming rule come back as problems.
func (l *Local) ListTicketFiles(id string) ([]TicketFile, []doc.Problem, error) {
	dir := IssuesDir(id)
	if err := l.checkPath(dir); err != nil {
		return nil, nil, err
	}
	entries, err := os.ReadDir(filepath.Join(l.Root, dir))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, apperr.New("read_failed", "can't list %s: %v", dir, err)
	}
	var files []TicketFile
	var problems []doc.Problem
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || !strings.HasSuffix(name, ".md") {
			continue
		}
		m := ticketFileRe.FindStringSubmatch(name)
		n, nerr := 0, error(nil)
		if m != nil {
			n, nerr = doc.ParseTicketNumber(m[1])
		}
		if m == nil || nerr != nil || e.IsDir() {
			problems = append(problems, doc.Problem{Code: "bad_ticket_file", Message: dir + "/" + name + " isn't named `<NN>-<slug>.md`"})
			continue
		}
		files = append(files, TicketFile{Number: n, Path: dir + "/" + name})
	}
	sort.Slice(files, func(i, j int) bool {
		if files[i].Number != files[j].Number {
			return files[i].Number < files[j].Number
		}
		return files[i].Path < files[j].Path
	})
	return files, problems, nil
}

// Exists reports whether a repo-relative path exists.
func (l *Local) Exists(rel string) (bool, error) {
	if err := l.checkPath(rel); err != nil {
		return false, err
	}
	_, err := os.Lstat(filepath.Join(l.Root, rel))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, apperr.New("read_failed", "can't read %s: %v", rel, err)
	}
	return true, nil
}

// Read returns a document's bytes. Missing files return a not_found error.
func (l *Local) Read(rel string) ([]byte, error) {
	if err := l.checkPath(rel); err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(l.Root, rel))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, apperr.New("not_found", "%s doesn't exist", rel)
	}
	if err != nil {
		return nil, apperr.New("read_failed", "can't read %s: %v", rel, err)
	}
	return b, nil
}

// Write atomically replaces or creates a document. With create set, it
// fails if the file already exists.
func (l *Local) Write(rel string, data []byte, create bool) error {
	if err := l.checkPath(rel); err != nil {
		return err
	}
	full := filepath.Join(l.Root, rel)
	dir := filepath.Dir(full)
	if err := l.mkdirs(filepath.Dir(rel)); err != nil {
		return err
	}
	mode := fs.FileMode(0o644)
	if info, err := os.Lstat(full); err == nil {
		if create {
			return apperr.New("already_exists", "%s already exists", rel)
		}
		mode = info.Mode().Perm()
	} else if !errors.Is(err, fs.ErrNotExist) {
		return apperr.New("write_failed", "can't inspect %s: %v", rel, err)
	} else if !create {
		return apperr.New("not_found", "%s doesn't exist", rel)
	}

	var suffix [6]byte
	_, _ = rand.Read(suffix[:])
	tmp := filepath.Join(dir, ".issue-tracker-"+hex.EncodeToString(suffix[:])+".tmp")
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return apperr.New("write_failed", "can't write %s: %v", rel, err)
	}
	cleanup := func(err error) error {
		f.Close()
		os.Remove(tmp)
		return apperr.New("write_failed", "can't write %s: %v", rel, err)
	}
	if _, err := f.Write(data); err != nil {
		return cleanup(err)
	}
	if err := f.Sync(); err != nil {
		return cleanup(err)
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return apperr.New("write_failed", "can't write %s: %v", rel, err)
	}
	if err := os.Chmod(tmp, mode); err != nil {
		os.Remove(tmp)
		return apperr.New("write_failed", "can't write %s: %v", rel, err)
	}
	if err := os.Rename(tmp, full); err != nil {
		os.Remove(tmp)
		return apperr.New("write_failed", "can't write %s: %v", rel, err)
	}
	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		d.Close()
	}
	return nil
}

func (l *Local) mkdirs(rel string) error {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	cur := l.Root
	for _, p := range parts {
		cur = filepath.Join(cur, p)
		info, err := os.Lstat(cur)
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.Mkdir(cur, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
				return apperr.New("write_failed", "can't create %s: %v", cur, err)
			}
			continue
		}
		if err != nil {
			return apperr.New("write_failed", "can't inspect %s: %v", cur, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 || !info.IsDir() {
			return apperr.New("unsafe_path", "%s is a symlink or not a directory", cur)
		}
	}
	return nil
}

// checkPath rejects paths that leave `.scratch/` or pass through a symlink.
func (l *Local) checkPath(rel string) error {
	clean := filepath.ToSlash(filepath.Clean(rel))
	if clean != ScratchDir && !strings.HasPrefix(clean, ScratchDir+"/") {
		return apperr.New("unsafe_path", "%s is outside %s", rel, ScratchDir)
	}
	cur := l.Root
	for _, p := range strings.Split(clean, "/") {
		if p == ".." || p == "." || p == "" {
			return apperr.New("unsafe_path", "%s is not a plain path", rel)
		}
		cur = filepath.Join(cur, p)
		info, err := os.Lstat(cur)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return apperr.New("read_failed", "can't inspect %s: %v", cur, err)
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return apperr.New("unsafe_path", "%s is a symlink; issue-tracker doesn't follow symlinks in %s", filepath.ToSlash(strings.TrimPrefix(cur, l.Root+string(filepath.Separator))), ScratchDir)
		}
	}
	return nil
}
