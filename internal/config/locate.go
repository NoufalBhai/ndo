package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NoufalBhai/ndo/internal/recipe"
)

// DefaultLocalFileName is the local recipe filename searched for upward
// from the current directory, overridable via NDO_LOCAL_FILE.
const DefaultLocalFileName = ".ndo.toml"

// ResolveNDOHome returns the directory holding central.toml and
// config.toml: NDO_HOME if set, otherwise ~/.ndo. No XDG fallback.
func ResolveNDOHome() (string, error) {
	if home := os.Getenv("NDO_HOME"); home != "" {
		return home, nil
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHome, ".ndo"), nil
}

// LocalFileName returns NDO_LOCAL_FILE if set, otherwise
// DefaultLocalFileName.
func LocalFileName() string {
	if name := os.Getenv("NDO_LOCAL_FILE"); name != "" {
		return name
	}
	return DefaultLocalFileName
}

// SearchUpward walks from startDir up through parent directories (like git
// discovering .git) looking for a file named filename. Returns its full
// path and true if found, or "" and false if it reaches the filesystem
// root without finding one.
func SearchUpward(startDir, filename string) (string, bool) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", false
	}
	for {
		candidate := filepath.Join(dir, filename)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// EnsureLocalFile returns the path to the nearest local recipe file found
// via upward search from cwd, creating an empty one in cwd if none exists
// anywhere in the tree. Shared by `ndo add --local` and `ndo init` so
// "create if missing" logic lives in one place.
func EnsureLocalFile(cwd string) (string, error) {
	if path, ok := SearchUpward(cwd, LocalFileName()); ok {
		return path, nil
	}
	path := filepath.Join(cwd, LocalFileName())
	if err := SaveRecipeFile(path, RecipeFile{Recipes: map[string]recipe.Recipe{}}); err != nil {
		return "", fmt.Errorf("creating %s: %w", path, err)
	}
	return path, nil
}

// EnsureNDOHome creates ndoHome (e.g. ~/.ndo) if it doesn't exist yet, so
// central.toml/config.toml can be written into it.
func EnsureNDOHome(ndoHome string) error {
	if err := os.MkdirAll(ndoHome, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", ndoHome, err)
	}
	return nil
}
