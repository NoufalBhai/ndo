// Package update implements ndo's self-update: detecting how the running
// binary was likely installed, and — only when there's no package manager
// that would be left out of sync — downloading the latest release,
// verifying its checksum, and replacing the running binary in place.
package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const repo = "green-threads/ndo"

// Overridable so tests can point these at an httptest.Server instead of
// the real GitHub endpoints.
var (
	apiBaseURL      = "https://api.github.com"
	downloadBaseURL = "https://github.com"
)

// Channel identifies how ndo was likely installed.
type Channel int

const (
	// Unknown means no package manager was detected — the safe case for
	// SelfReplace, since there's nothing else tracking this binary.
	Unknown Channel = iota
	Homebrew
	Scoop
	SystemPackage // .deb/.rpm, installed to /usr/bin by nfpm
	GoInstall
)

// toSlash normalizes path separators to "/" regardless of the OS this
// process is actually running on — unlike filepath.ToSlash, which only
// converts filepath.Separator for the *host* OS, and so is a no-op for
// backslash-separated Windows-style paths when running on Linux/macOS
// (as happens here: DetectChannel takes an explicit goos so its tests can
// exercise Windows-shaped paths from any CI runner).
func toSlash(p string) string {
	return strings.ReplaceAll(p, `\`, "/")
}

// DetectChannel guesses the install channel from the running executable's
// path and a snapshot of the environment, using the layout each real
// install path leaves behind: Homebrew's Cellar, Scoop's apps directory,
// the /usr/bin bindir nfpm uses for the .deb/.rpm packages, and Go's
// GOBIN/GOPATH. Kept pure — no direct env/OS reads — so it's unit tested
// without touching the real filesystem or environment.
func DetectChannel(execPath, goos string, env map[string]string) Channel {
	lower := strings.ToLower(toSlash(execPath))

	if strings.Contains(lower, "/cellar/") || strings.Contains(lower, "linuxbrew") {
		return Homebrew
	}
	if goos == "windows" && strings.Contains(lower, "/scoop/") {
		return Scoop
	}

	if goBin := env["GOBIN"]; goBin != "" {
		if strings.HasPrefix(lower, strings.ToLower(toSlash(goBin))) {
			return GoInstall
		}
	}
	if goPath := env["GOPATH"]; goPath != "" {
		if strings.HasPrefix(lower, strings.ToLower(toSlash(goPath))+"/bin") {
			return GoInstall
		}
	}

	// nfpm's bindir (.goreleaser.yaml) is exactly /usr/bin — distinct from
	// install.sh's /usr/local/bin (or ~/.local/bin) fallback, so this
	// specific path reliably means "installed via .deb/.rpm", not
	// "installed via the install script".
	if goos != "windows" && lower == "/usr/bin/ndo" {
		return SystemPackage
	}

	return Unknown
}

// InstructionFor returns the right upgrade command for a detected
// channel, and whether it's safe for SelfReplace to run instead
// (selfReplace is true only for Unknown — no package manager to desync).
func InstructionFor(ch Channel) (instruction string, selfReplace bool) {
	switch ch {
	case Homebrew:
		return "brew upgrade ndo", false
	case Scoop:
		return "scoop update ndo", false
	case SystemPackage:
		return "your system package manager, e.g. `sudo apt install --only-upgrade ndo` or `sudo dnf upgrade ndo`", false
	case GoInstall:
		return "go install github.com/green-threads/ndo/cmd/ndo@latest", false
	default:
		return "", true
	}
}

type releaseInfo struct {
	TagName string `json:"tag_name"`
}

// LatestRelease returns the tag name of the latest GitHub release.
func LatestRelease(client *http.Client) (tag string, err error) {
	req, err := http.NewRequest(http.MethodGet, apiBaseURL+"/repos/"+repo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("checking latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checking latest release: unexpected status %s", resp.Status)
	}

	var info releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("parsing latest release response: %w", err)
	}
	if info.TagName == "" {
		return "", fmt.Errorf("latest release response had no tag_name")
	}
	return info.TagName, nil
}

// AssetName builds the release asset filename for goos/goarch/tag,
// matching the naming goreleaser uses (and install.sh parses).
func AssetName(goos, goarch, tag string) string {
	version := strings.TrimPrefix(tag, "v")
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("ndo_%s_%s_%s.%s", version, goos, goarch, ext)
}

// VerifyChecksum checks archiveData's sha256 against the entry for
// assetName in a checksums.txt (sha256sum format: "<hex>  <filename>").
func VerifyChecksum(archiveData, checksumsTxt []byte, assetName string) error {
	sum := sha256.Sum256(archiveData)
	got := hex.EncodeToString(sum[:])

	for _, line := range strings.Split(string(checksumsTxt), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[1] != assetName {
			continue
		}
		if fields[0] != got {
			return fmt.Errorf("checksum mismatch for %s: got %s, want %s", assetName, got, fields[0])
		}
		return nil
	}
	return fmt.Errorf("no checksum entry found for %s", assetName)
}

// SelfReplace downloads tag's release asset for goos/goarch, verifies it
// against checksums.txt, extracts the ndo binary, and atomically replaces
// the executable at execPath with it.
func SelfReplace(client *http.Client, execPath, tag, goos, goarch string) error {
	assetName := AssetName(goos, goarch, tag)

	archiveData, err := download(client, releaseAssetURL(tag, assetName))
	if err != nil {
		return fmt.Errorf("downloading %s: %w", assetName, err)
	}
	checksumsData, err := download(client, releaseAssetURL(tag, "checksums.txt"))
	if err != nil {
		return fmt.Errorf("downloading checksums.txt: %w", err)
	}
	if err := VerifyChecksum(archiveData, checksumsData, assetName); err != nil {
		return err
	}

	binaryName := "ndo"
	extract := ExtractTarGz
	if goos == "windows" {
		binaryName = "ndo.exe"
		extract = ExtractZip
	}
	binary, err := extract(archiveData, binaryName)
	if err != nil {
		return err
	}

	info, err := os.Stat(execPath)
	if err != nil {
		return fmt.Errorf("stat-ing current executable: %w", err)
	}
	return ReplaceExecutable(execPath, binary, info.Mode())
}

func releaseAssetURL(tag, assetName string) string {
	return fmt.Sprintf("%s/%s/releases/download/%s/%s", downloadBaseURL, repo, tag, assetName)
}

func download(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s for %s", resp.Status, url)
	}
	return io.ReadAll(resp.Body)
}

// ExtractTarGz returns binaryName's contents from a .tar.gz archive.
func ExtractTarGz(data []byte, binaryName string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("opening gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}
		if filepath.Base(hdr.Name) == binaryName {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("%s not found in archive", binaryName)
}

// ExtractZip returns binaryName's contents from a .zip archive.
func ExtractZip(data []byte, binaryName string) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("opening zip: %w", err)
	}
	for _, f := range r.File {
		if filepath.Base(f.Name) != binaryName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("opening %s in archive: %w", binaryName, err)
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, fmt.Errorf("%s not found in archive", binaryName)
}
