//go:build linux || darwin

package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ehutchllew/dotfiles/tools/issue-tracker/internal/apperr"
)

// lockTimeout is how long a command waits for a feature lock.
const lockTimeout = 30 * time.Second

func defaultLockDir() string {
	if dir := os.Getenv("ISSUE_TRACKER_LOCK_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("issue-tracker-%d", os.Getuid()))
}

// Lock takes a per-feature lock outside the working tree. Readers share it,
// and writers hold it alone. Call the returned function to release it.
func (l *Local) Lock(feature string, exclusive bool) (func(), error) {
	if err := os.MkdirAll(l.LockDir, 0o700); err != nil {
		return nil, apperr.New("lock_failed", "can't create lock directory %s: %v", l.LockDir, err)
	}
	sum := sha256.Sum256([]byte(l.Canonical + "\x00" + feature))
	path := filepath.Join(l.LockDir, hex.EncodeToString(sum[:16])+".lock")
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, apperr.New("lock_failed", "can't open lock %s: %v", path, err)
	}
	how := syscall.LOCK_SH
	if exclusive {
		how = syscall.LOCK_EX
	}
	deadline := time.Now().Add(lockTimeout)
	for {
		err := syscall.Flock(int(f.Fd()), how|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EINTR) {
			f.Close()
			return nil, apperr.New("lock_failed", "can't lock %s: %v", path, err)
		}
		if time.Now().After(deadline) {
			f.Close()
			return nil, apperr.New("lock_timeout", "another issue-tracker command held feature %q for over %s", feature, lockTimeout)
		}
		time.Sleep(20 * time.Millisecond)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}
