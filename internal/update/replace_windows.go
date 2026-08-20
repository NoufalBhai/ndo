//go:build windows

package update

import (
	"fmt"
	"os"
)

// ReplaceExecutable replaces the file at path with newBinary's contents.
// Windows won't allow overwriting or deleting a running .exe directly,
// but *renaming* it is allowed — so the current binary is renamed to a
// ".old" sibling first, the new one takes its place, and the ".old" file
// is removed on a best-effort basis (it may still be locked by this very
// process, in which case the next update run cleans it up before doing
// anything else).
func ReplaceExecutable(path string, newBinary []byte, mode os.FileMode) error {
	oldPath := path + ".old"
	_ = os.Remove(oldPath) // leftover from a previous update; fine if absent/locked

	if err := os.Rename(path, oldPath); err != nil {
		return fmt.Errorf("moving current binary aside: %w", err)
	}

	if err := os.WriteFile(path, newBinary, mode); err != nil {
		// Best-effort: restore the original so the install isn't left broken.
		_ = os.Rename(oldPath, path)
		return fmt.Errorf("writing new binary: %w", err)
	}

	_ = os.Remove(oldPath) // best-effort; a lock here just means it waits for next time
	return nil
}
