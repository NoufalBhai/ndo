package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectChannel(t *testing.T) {
	tests := []struct {
		name string
		path string
		goos string
		env  map[string]string
		want Channel
	}{
		{"homebrew macos", "/usr/local/Cellar/ndo/0.3.1/bin/ndo", "darwin", nil, Homebrew},
		{"homebrew apple silicon", "/opt/homebrew/Cellar/ndo/0.3.1/bin/ndo", "darwin", nil, Homebrew},
		{"linuxbrew", "/home/linuxbrew/.linuxbrew/bin/ndo", "linux", nil, Homebrew},
		{"scoop", `C:\Users\alex\scoop\apps\ndo\current\ndo.exe`, "windows", nil, Scoop},
		{"go install via GOBIN", "/home/alex/gobin/ndo", "linux", map[string]string{"GOBIN": "/home/alex/gobin"}, GoInstall},
		{"go install via GOPATH", "/home/alex/go/bin/ndo", "linux", map[string]string{"GOPATH": "/home/alex/go"}, GoInstall},
		{"deb/rpm bindir", "/usr/bin/ndo", "linux", nil, SystemPackage},
		{"install.sh default location", "/usr/local/bin/ndo", "linux", nil, Unknown},
		{"install.sh fallback location", "/home/alex/.local/bin/ndo", "linux", nil, Unknown},
		{"raw exe anywhere on windows", `C:\tools\ndo.exe`, "windows", nil, Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectChannel(tt.path, tt.goos, tt.env); got != tt.want {
				t.Fatalf("DetectChannel(%q, %q, %v) = %v, want %v", tt.path, tt.goos, tt.env, got, tt.want)
			}
		})
	}
}

func TestInstructionForOnlyUnknownAllowsSelfReplace(t *testing.T) {
	for _, ch := range []Channel{Homebrew, Scoop, SystemPackage, GoInstall} {
		if _, selfReplace := InstructionFor(ch); selfReplace {
			t.Errorf("InstructionFor(%v) selfReplace = true, want false — must never self-replace a package-manager-owned binary", ch)
		}
	}
	if instruction, selfReplace := InstructionFor(Unknown); !selfReplace || instruction != "" {
		t.Errorf("InstructionFor(Unknown) = (%q, %v), want (\"\", true)", instruction, selfReplace)
	}
}

func TestAssetName(t *testing.T) {
	if got, want := AssetName("linux", "amd64", "v1.2.3"), "ndo_1.2.3_linux_amd64.tar.gz"; got != want {
		t.Fatalf("AssetName() = %q, want %q", got, want)
	}
	if got, want := AssetName("windows", "amd64", "v1.2.3"), "ndo_1.2.3_windows_amd64.zip"; got != want {
		t.Fatalf("AssetName() = %q, want %q", got, want)
	}
}

func TestVerifyChecksum(t *testing.T) {
	data := []byte("archive contents")
	sum := sha256.Sum256(data)
	hexSum := hex.EncodeToString(sum[:])
	checksums := []byte(fmt.Sprintf("%s  ndo_1.0.0_linux_amd64.tar.gz\nsomeotherhash  otherfile.tar.gz\n", hexSum))

	if err := VerifyChecksum(data, checksums, "ndo_1.0.0_linux_amd64.tar.gz"); err != nil {
		t.Fatalf("VerifyChecksum() error = %v, want nil", err)
	}
	if err := VerifyChecksum([]byte("tampered"), checksums, "ndo_1.0.0_linux_amd64.tar.gz"); err == nil {
		t.Fatal("VerifyChecksum() = nil for tampered data, want a mismatch error")
	}
	if err := VerifyChecksum(data, checksums, "nonexistent.tar.gz"); err == nil {
		t.Fatal("VerifyChecksum() = nil for an asset with no checksum entry, want an error")
	}
}

func TestExtractTarGz(t *testing.T) {
	want := []byte("#!/fake binary contents")
	data := buildTarGz(t, "ndo", want)

	got, err := ExtractTarGz(data, "ndo")
	if err != nil {
		t.Fatalf("ExtractTarGz() error: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("ExtractTarGz() = %q, want %q", got, want)
	}

	if _, err := ExtractTarGz(data, "nope"); err == nil {
		t.Fatal("ExtractTarGz() = nil error for a missing member, want an error")
	}
}

func TestExtractZip(t *testing.T) {
	want := []byte("MZ fake exe contents")
	data := buildZip(t, "ndo.exe", want)

	got, err := ExtractZip(data, "ndo.exe")
	if err != nil {
		t.Fatalf("ExtractZip() error: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("ExtractZip() = %q, want %q", got, want)
	}

	if _, err := ExtractZip(data, "nope.exe"); err == nil {
		t.Fatal("ExtractZip() = nil error for a missing member, want an error")
	}
}

func TestReplaceExecutableSwapsContentAndPreservesPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fake-ndo")
	if err := os.WriteFile(path, []byte("old contents"), 0o755); err != nil {
		t.Fatal(err)
	}

	newContents := []byte("new contents")
	if err := ReplaceExecutable(path, newContents, 0o755); err != nil {
		t.Fatalf("ReplaceExecutable() error: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, newContents) {
		t.Fatalf("file content = %q, want %q", got, newContents)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("directory has %d entries after replace, want exactly 1 (no leftover temp/.old files): %v", len(entries), entries)
	}
}

func TestLatestRelease(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/green-threads/ndo/releases/latest" {
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
		w.Write([]byte(`{"tag_name": "v9.9.9"}`))
	}))
	defer srv.Close()

	restore := apiBaseURL
	apiBaseURL = srv.URL
	defer func() { apiBaseURL = restore }()

	tag, err := LatestRelease(srv.Client())
	if err != nil {
		t.Fatalf("LatestRelease() error: %v", err)
	}
	if tag != "v9.9.9" {
		t.Fatalf("LatestRelease() = %q, want %q", tag, "v9.9.9")
	}
}

// TestSelfReplaceEndToEnd exercises the full download -> verify -> extract
// -> replace pipeline against an httptest.Server standing in for GitHub,
// and a throwaway temp file standing in for the running executable —
// never touching the real network or a real binary.
func TestSelfReplaceEndToEnd(t *testing.T) {
	binaryContents := []byte("new ndo binary contents")
	assetName := AssetName("linux", "amd64", "v1.2.3")
	archiveData := buildTarGz(t, "ndo", binaryContents)
	sum := sha256.Sum256(archiveData)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), assetName)

	mux := http.NewServeMux()
	mux.HandleFunc("/green-threads/ndo/releases/download/v1.2.3/"+assetName, func(w http.ResponseWriter, r *http.Request) {
		w.Write(archiveData)
	})
	mux.HandleFunc("/green-threads/ndo/releases/download/v1.2.3/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(checksums))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	restore := downloadBaseURL
	downloadBaseURL = srv.URL
	defer func() { downloadBaseURL = restore }()

	execPath := filepath.Join(t.TempDir(), "fake-ndo")
	if err := os.WriteFile(execPath, []byte("old ndo binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := SelfReplace(srv.Client(), execPath, "v1.2.3", "linux", "amd64"); err != nil {
		t.Fatalf("SelfReplace() error: %v", err)
	}

	got, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, binaryContents) {
		t.Fatalf("executable content after SelfReplace = %q, want %q", got, binaryContents)
	}
}

func TestSelfReplaceRejectsTamperedArchive(t *testing.T) {
	assetName := AssetName("linux", "amd64", "v1.2.3")
	archiveData := buildTarGz(t, "ndo", []byte("legit contents"))

	mux := http.NewServeMux()
	mux.HandleFunc("/green-threads/ndo/releases/download/v1.2.3/"+assetName, func(w http.ResponseWriter, r *http.Request) {
		w.Write(archiveData)
	})
	mux.HandleFunc("/green-threads/ndo/releases/download/v1.2.3/checksums.txt", func(w http.ResponseWriter, r *http.Request) {
		// Checksum for different content than what's actually served.
		sum := sha256.Sum256([]byte("something else entirely"))
		w.Write([]byte(fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), assetName)))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	restore := downloadBaseURL
	downloadBaseURL = srv.URL
	defer func() { downloadBaseURL = restore }()

	execPath := filepath.Join(t.TempDir(), "fake-ndo")
	original := []byte("original contents, must survive")
	if err := os.WriteFile(execPath, original, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := SelfReplace(srv.Client(), execPath, "v1.2.3", "linux", "amd64"); err == nil {
		t.Fatal("SelfReplace() = nil error for a checksum mismatch, want an error")
	}

	got, err := os.ReadFile(execPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("executable was modified despite a checksum mismatch: got %q, want untouched %q", got, original)
	}
}

func buildTarGz(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func buildZip(t *testing.T, name string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	f, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
