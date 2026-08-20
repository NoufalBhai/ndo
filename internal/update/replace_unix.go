//go:build !windows

package update

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReplaceExecutable atomically replaces the file at path with newBinary's
// contents. On Unix, os.Rename within the same directory is atomic, so
// the new binary is written to a sibling temp file first and renamed into
// place — any process still holding the old inode open (including this
// one, until it exits) keeps working fine.
func ReplaceExecutable(path string, newBinary []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".ndo-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file in %s: %w", dir, err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // no-op once successfully renamed away

	if _, err := tmp.Write(newBinary); err != nil {
		tmp.Close()
		return fmt.Errorf("writing new binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		return fmt.Errorf("setting permissions on new binary: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}
