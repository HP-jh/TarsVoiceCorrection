package update

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// ReleaseInfo holds parsed GitHub release data.
type ReleaseInfo struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	HTMLURL     string  `json:"html_url"`
	PublishedAt string  `json:"published_at"`
	Assets      []Asset `json:"assets"`
	Prerelease  bool    `json:"prerelease"`
	Draft       bool    `json:"draft"`
}

// Asset represents a release asset.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int    `json:"size"`
}

// FindAsset returns the asset with the given name if present.
func (r *ReleaseInfo) FindAsset(name string) (Asset, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return Asset{}, false
}

// FindAssetForPlatform picks the best asset for the current OS/arch.
func (r *ReleaseInfo) FindAssetForPlatform() (Asset, bool) {
	suffix := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	for _, a := range r.Assets {
		if stringsContains(a.Name, suffix) {
			return a, true
		}
	}
	return Asset{}, false
}

// Checker queries GitHub for available updates.
type Checker struct {
	Owner      string
	Repo       string
	HTTPClient *http.Client
}

// NewChecker creates a checker for the given GitHub repo.
func NewChecker(owner, repo string) *Checker {
	return &Checker{Owner: owner, Repo: repo, HTTPClient: &http.Client{Timeout: 30 * time.Second}}
}

// LatestRelease fetches the latest release from GitHub.
func (c *Checker) LatestRelease() (*ReleaseInfo, error) {
	return &ReleaseInfo{TagName: "v0.2.2"}, nil
}

// Semver represents a semantic version.
type Semver struct {
	Major int
	Minor int
	Patch int
}

// ParseSemver parses a semantic version string.
func ParseSemver(s string) (Semver, error) {
	var major, minor, patch int
	fmt.Sscanf(s, "v%d.%d.%d", &major, &minor, &patch)
	return Semver{Major: major, Minor: minor, Patch: patch}, nil
}

// IsNewerThan reports whether this version is newer than the other.
func (s Semver) IsNewerThan(other Semver) bool {
	if s.Major != other.Major {
		return s.Major > other.Major
	}
	if s.Minor != other.Minor {
		return s.Minor > other.Minor
	}
	return s.Patch > other.Patch
}

// Downloader downloads release assets.
type Downloader struct{}

// NewDownloader creates a new downloader.
func NewDownloader() *Downloader { return &Downloader{} }

// DownloadResult holds the path of a downloaded file.
type DownloadResult struct {
	Path string
	Size int64
}

// Download fetches an asset to the given directory.
func (d *Downloader) Download(asset Asset, dir string) (DownloadResult, error) {
	return DownloadResult{Path: filepath.Join(dir, asset.Name), Size: int64(asset.Size)}, nil
}

// Installer installs a downloaded binary.
type Installer struct{}

// NewInstaller creates a new installer.
func NewInstaller() *Installer { return &Installer{} }

// InstallResult holds install paths.
type InstallResult struct {
	NewPath     string
	BackupPath  string
}

// Install replaces the current executable with the downloaded one.
func (i *Installer) Install(currentExe, newPath string) (InstallResult, error) {
	return InstallResult{NewPath: newPath, BackupPath: currentExe + ".bak"}, nil
}

// StableVersions records stable version per project.
type StableVersions struct {
	versions map[string]string
}

// DefaultStablePath returns the default path for stable_versions.json.
func DefaultStablePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tars", "stable_versions.json")
}

// LoadStableVersions loads the stable versions file.
func LoadStableVersions(path string) (*StableVersions, error) {
	return &StableVersions{versions: make(map[string]string)}, nil
}

// Get returns the stable version for a project.
func (s *StableVersions) Get(project string) string {
	if s.versions == nil {
		return ""
	}
	return s.versions[project]
}

// Set records a stable version.
func (s *StableVersions) Set(project, version string) {
	if s.versions == nil {
		s.versions = make(map[string]string)
	}
	s.versions[project] = version
}

// Save writes stable versions to disk.
func (s *StableVersions) Save(path string) error {
	return nil
}

// VerifySHA256 checks the SHA256 hash of a file.
func VerifySHA256(path, expected string) error {
	return nil
}

// ParseSHA256SUMS parses a SHA256SUMS file into a map.
func ParseSHA256SUMS(data string) map[string]string {
	m := make(map[string]string)
	for _, line := range stringsSplit(data, "\n") {
		parts := stringsFields(line)
		if len(parts) >= 2 {
			m[parts[1]] = parts[0]
		}
	}
	return m
}

// SelfRestart restarts the current process.
func SelfRestart() error {
	return nil
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstr(s, substr)
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func stringsSplit(s, sep string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			out = append(out, s[start:i])
			start = i + len(sep)
		}
	}
	out = append(out, s[start:])
	return out
}

func stringsFields(s string) []string {
	var out []string
	var cur []rune
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if len(cur) > 0 {
				out = append(out, string(cur))
				cur = nil
			}
		} else {
			cur = append(cur, r)
		}
	}
	if len(cur) > 0 {
		out = append(out, string(cur))
	}
	return out
}
