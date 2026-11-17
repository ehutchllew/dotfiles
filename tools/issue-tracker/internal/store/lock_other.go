//go:build !linux && !darwin

package store

import (
	"os"
	"path/filepath"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
)

func defaultLockDir() string {
	if dir := os.Getenv("ISSUE_TRACKER_LOCK_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(os.TempDir(), "issue-tracker")
}

// Lock is unavailable outside Linux and macOS.
func (l *Local) Lock(feature string, exclusive bool) (func(), error) {
	return nil, apperr.New("lock_failed", "feature locks are supported only on Linux and macOS")
}
