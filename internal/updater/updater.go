package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Version is the current application version (set during build or hardcoded)
var Version = "1.0.0"

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// UpdateInfo contains information about an available update
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseNotes   string `json:"releaseNotes"`
	DownloadURL    string `json:"downloadURL"`
	FileName       string `json:"fileName"`
	FileSize       int64  `json:"fileSize"`
}

// Updater handles application updates
type Updater struct {
	Owner      string
	Repo       string
	CurrentVer string
}

// NewUpdater creates a new updater instance
func NewUpdater(owner, repo string) *Updater {
	return &Updater{
		Owner:      owner,
		Repo:       repo,
		CurrentVer: Version,
	}
}

// CheckForUpdates checks GitHub for new releases
func (u *Updater) CheckForUpdates() (*UpdateInfo, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.Owner, u.Repo)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return &UpdateInfo{
			Available:      false,
			CurrentVersion: u.CurrentVer,
			LatestVersion:  u.CurrentVer,
		}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %v", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(u.CurrentVer, "v")

	info := &UpdateInfo{
		Available:      compareVersions(latestVersion, currentVersion) > 0,
		CurrentVersion: u.CurrentVer,
		LatestVersion:  release.TagName,
		ReleaseNotes:   release.Body,
	}

	// Find the appropriate asset for current platform
	assetName := u.getAssetName()
	for _, asset := range release.Assets {
		if strings.Contains(strings.ToLower(asset.Name), strings.ToLower(assetName)) {
			info.DownloadURL = asset.BrowserDownloadURL
			info.FileName = asset.Name
			info.FileSize = asset.Size
			break
		}
	}

	return info, nil
}

// DownloadUpdate downloads the update to a temporary location
func (u *Updater) DownloadUpdate(url string, onProgress func(downloaded, total int64)) (string, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download update: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, "autologcollector-update.exe")

	out, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer out.Close()

	// Download with progress
	total := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return "", fmt.Errorf("failed to write update file: %v", writeErr)
			}
			downloaded += int64(n)
			if onProgress != nil {
				onProgress(downloaded, total)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("download error: %v", err)
		}
	}

	return tempFile, nil
}

// ApplyUpdate replaces the current executable with the new one
func (u *Updater) ApplyUpdate(newExePath string) error {
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %v", err)
	}

	currentExe, err = filepath.EvalSymlinks(currentExe)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %v", err)
	}

	// Backup current executable
	backupPath := currentExe + ".old"
	_ = os.Remove(backupPath) // Remove old backup if exists

	// Rename current to backup
	if err := os.Rename(currentExe, backupPath); err != nil {
		return fmt.Errorf("failed to backup current executable: %v", err)
	}

	// Move new executable to current location
	if err := os.Rename(newExePath, currentExe); err != nil {
		// Try to restore backup
		_ = os.Rename(backupPath, currentExe)
		return fmt.Errorf("failed to install update: %v", err)
	}

	return nil
}

// GetCurrentVersion returns the current version
func (u *Updater) GetCurrentVersion() string {
	return u.CurrentVer
}

// getAssetName returns the expected asset name for current platform
func (u *Updater) getAssetName() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	default:
		return runtime.GOOS
	}
}

// compareVersions compares two version strings
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}

	return 0
}
