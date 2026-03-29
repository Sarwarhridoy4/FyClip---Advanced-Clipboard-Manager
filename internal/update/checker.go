// File: internal/update/checker.go
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Sarwarhridoy4/FyClip---Advanced-Clipboard-Manager/internal/logger"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName    string       `json:"tag_name"`
	Name       string       `json:"name"`
	Body       string       `json:"body"`
	Prerelease bool         `json:"prerelease"`
	PublishedAt string     `json:"published_at"`
	Assets     []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a release asset
type GitHubAsset struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"browser_download_url"`
	ContentType string `json:"content_type"`
}

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	ReleaseNotes   string
	DownloadURL    string
	AssetName      string
	AssetSize      int64
	IsPrerelease   bool
}

// Checker handles update checking functionality
type Checker struct {
	owner           string
	repo            string
	currentVersion  string
	httpClient      *http.Client
	downloadTimeout time.Duration
	log             *logger.Logger
}

// Option is a functional option for Checker
type Option func(*Checker)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) Option {
	return func(c *Checker) {
		c.httpClient = client
	}
}

// WithDownloadTimeout sets the download timeout
func WithDownloadTimeout(timeout time.Duration) Option {
	return func(c *Checker) {
		c.downloadTimeout = timeout
	}
}

// WithLogger sets a custom logger
func WithLogger(log *logger.Logger) Option {
	return func(c *Checker) {
		c.log = log
	}
}

// NewChecker creates a new update checker
func NewChecker(owner, repo, currentVersion string, opts ...Option) *Checker {
	c := &Checker{
		owner:           owner,
		repo:            repo,
		currentVersion:  currentVersion,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
		downloadTimeout: 5 * time.Minute,
		log:             logger.Get(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// CheckForUpdate checks if a new version is available
func (c *Checker) CheckForUpdate(ctx context.Context) (*UpdateInfo, error) {
	c.log.Info("Checking for updates...")

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", c.owner, c.repo)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "FyClip-Update-Checker")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}

	// Find the appropriate asset for the current platform
	asset := c.findAsset(release.Assets)
	if asset == nil {
		return nil, fmt.Errorf("no compatible asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	releaseURL := fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", c.owner, c.repo, release.TagName)

	updateInfo := &UpdateInfo{
		CurrentVersion: c.currentVersion,
		LatestVersion:  strings.TrimPrefix(release.TagName, "v"),
		ReleaseURL:     releaseURL,
		ReleaseNotes:   release.Body,
		DownloadURL:    asset.DownloadURL,
		AssetName:      asset.Name,
		AssetSize:      asset.Size,
		IsPrerelease:   release.Prerelease,
	}

	c.log.Info(fmt.Sprintf("Current version: %s, Latest version: %s", c.currentVersion, updateInfo.LatestVersion))

	return updateInfo, nil
}

// findAsset finds the appropriate asset for the current platform
func (c *Checker) findAsset(assets []GitHubAsset) *GitHubAsset {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map runtime GOOS to asset naming conventions
	osMap := map[string][]string{
		"linux":  {"linux", "Linux", "deb", "AppImage"},
		"windows": {"windows", "Windows", "exe", "msi"},
		"darwin":  {"darwin", "Darwin", "macos", "macOS", "dmg"},
	}

	archMap := map[string][]string{
		"amd64": {"amd64", "x86_64", "x64"},
		"arm64": {"arm64", "aarch64", "arm64"},
		"386":   {"386", "i386", "i686"},
	}

	var candidates []GitHubAsset

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)

		// Check OS match
		osMatch := false
		for _, os := range osMap[goos] {
			if strings.Contains(name, strings.ToLower(os)) {
				osMatch = true
				break
			}
		}
		if !osMatch {
			continue
		}

		// Check arch match
		archMatch := false
		for _, arch := range archMap[goarch] {
			if strings.Contains(name, strings.ToLower(arch)) {
				archMatch = true
				break
			}
		}
		if !archMatch {
			continue
		}

		candidates = append(candidates, asset)
	}

	if len(candidates) == 0 {
		return nil
	}

	// Prefer .deb on Linux, .exe on Windows, .dmg on macOS
	preferredExt := map[string]string{
		"linux":   ".deb",
		"windows": ".exe",
		"darwin":  ".dmg",
	}

	if ext := preferredExt[goos]; ext != "" {
		for _, asset := range candidates {
			if strings.HasSuffix(asset.Name, ext) {
				return &asset
			}
		}
	}

	// Return first candidate
	return &candidates[0]
}

// IsUpdateAvailable checks if an update is available
func (c *Checker) IsUpdateAvailable(ctx context.Context) (bool, error) {
	updateInfo, err := c.CheckForUpdate(ctx)
	if err != nil {
		return false, err
	}

	return CompareVersions(updateInfo.LatestVersion, c.currentVersion) > 0, nil
}

// CompareVersions compares two semantic versions
// Returns: 1 if v1 > v2, 0 if v1 == v2, -1 if v1 < v2
func CompareVersions(v1, v2 string) int {
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	// Handle pre-release versions: 1.0.0-beta < 1.0.0
	v1IsPrerelease := strings.Contains(v1, "-")
	v2IsPrerelease := strings.Contains(v2, "-")

	// If both are pre-release or both are not, compare normally
	// If one is pre-release and other is not, non-prerelease is greater
	if v1IsPrerelease && !v2IsPrerelease {
		return -1
	}
	if !v1IsPrerelease && v2IsPrerelease {
		return 1
	}

	// Strip pre-release suffix for numeric comparison
	if v1IsPrerelease {
		v1 = strings.Split(v1, "-")[0]
	}
	if v2IsPrerelease {
		v2 = strings.Split(v2, "-")[0]
	}

	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(v1Parts) {
			fmt.Sscanf(v1Parts[i], "%d", &n1)
		}
		if i < len(v2Parts) {
			fmt.Sscanf(v2Parts[i], "%d", &n2)
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

// Downloader handles downloading and installing updates
type Downloader struct {
	checker       *Checker
	updateInfo    *UpdateInfo
	downloadPath  string
	log           *logger.Logger
}

// NewDownloader creates a new downloader
func NewDownloader(checker *Checker, updateInfo *UpdateInfo) *Downloader {
	return &Downloader{
		checker:      checker,
		updateInfo:   updateInfo,
		downloadPath: filepath.Join(os.TempDir(), updateInfo.AssetName),
		log:          logger.Get(),
	}
}

// Download downloads the update asset
func (d *Downloader) Download(ctx context.Context, progressFunc func(downloaded, total int64)) error {
	d.log.Info(fmt.Sprintf("Downloading %s...", d.updateInfo.AssetName))

	req, err := http.NewRequestWithContext(ctx, "GET", d.updateInfo.DownloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "FyClip-Update-Downloader")

	resp, err := d.checker.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	totalSize := resp.ContentLength
	if totalSize <= 0 {
		totalSize = d.updateInfo.AssetSize
	}

	// Create file
	file, err := os.Create(d.downloadPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Download with progress
	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := file.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write: %w", writeErr)
			}
			downloaded += int64(n)
			if progressFunc != nil {
				progressFunc(downloaded, totalSize)
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read error: %w", err)
		}
	}

	d.log.Info("Download complete")
	return nil
}

// GetDownloadPath returns the path to the downloaded file
func (d *Downloader) GetDownloadPath() string {
	return d.downloadPath
}

// Installer handles installing the update
type Installer struct {
	downloadPath  string
	appName       string
	log           *logger.Logger
	output        strings.Builder
}

// NewInstaller creates a new installer
func NewInstaller(downloadPath, appName string) *Installer {
	return &Installer{
		downloadPath: downloadPath,
		appName:      appName,
		log:          logger.Get(),
	}
}

// GetOutput returns the captured installation output
func (i *Installer) GetOutput() string {
	return i.output.String()
}

// Install installs the update based on the platform and file type
func (i *Installer) Install() error {
	filename := filepath.Base(i.downloadPath)
	ext := strings.ToLower(filepath.Ext(filename))

	i.log.Info(fmt.Sprintf("Installing %s...", filename))

	switch runtime.GOOS {
	case "linux":
		return i.installLinux(ext)
	case "windows":
		return i.installWindows(ext)
	case "darwin":
		return i.installDarwin(ext)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// installLinux handles Linux installation
func (i *Installer) installLinux(ext string) error {
	switch ext {
	case ".deb":
		return i.installDeb()
	case ".AppImage", ".appimage":
		return i.installAppImage()
	default:
		return fmt.Errorf("unsupported Linux package format: %s", ext)
	}
}

// installDeb installs a .deb package
func (i *Installer) installDeb() error {
	// Try pkexec first (graphical sudo), fall back to sudo
	var cmd *exec.Cmd
	if _, err := exec.LookPath("pkexec"); err == nil {
		cmd = exec.Command("pkexec", "dpkg", "-i", i.downloadPath)
	} else {
		cmd = exec.Command("sudo", "dpkg", "-i", i.downloadPath)
	}
	cmd.Stdout = &i.output
	cmd.Stderr = &i.output

	i.log.Info("Installing .deb package with dpkg...")
	err := cmd.Run()
	if err != nil {
		i.log.Error(fmt.Sprintf("dpkg installation failed: %v", err))
		i.log.Error(fmt.Sprintf("Installation output: %s", i.output.String()))
	}
	return err
}

// installAppImage makes an AppImage executable
func (i *Installer) installAppImage() error {
	// Make executable
	err := os.Chmod(i.downloadPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to make executable: %w", err)
	}

	i.log.Info("AppImage is ready to run. You can execute it directly.")
	i.log.Info(fmt.Sprintf("Location: %s", i.downloadPath))

	return nil
}

// installWindows handles Windows installation
func (i *Installer) installWindows(ext string) error {
	switch ext {
	case ".exe":
		return i.installExe()
	case ".msi":
		return i.installMsi()
	default:
		return fmt.Errorf("unsupported Windows package format: %s", ext)
	}
}

// installExe runs the installer executable
func (i *Installer) installExe() error {
	cmd := exec.Command(i.downloadPath, "/S") // Silent install
	cmd.Stdout = &i.output
	cmd.Stderr = &i.output

	i.log.Info("Running installer...")
	err := cmd.Run()
	if err != nil {
		i.log.Error(fmt.Sprintf("Installer execution failed: %v", err))
		i.log.Error(fmt.Sprintf("Installation output: %s", i.output.String()))
	}
	return err
}

// installMsi installs an MSI package
func (i *Installer) installMsi() error {
	cmd := exec.Command("msiexec", "/i", i.downloadPath, "/quiet")
	cmd.Stdout = &i.output
	cmd.Stderr = &i.output

	i.log.Info("Installing MSI package...")
	err := cmd.Run()
	if err != nil {
		i.log.Error(fmt.Sprintf("MSI installation failed: %v", err))
		i.log.Error(fmt.Sprintf("Installation output: %s", i.output.String()))
	}
	return err
}

// installDarwin handles macOS installation
func (i *Installer) installDarwin(ext string) error {
	switch ext {
	case ".dmg":
		return i.installDmg()
	case ".zip":
		return i.installZip()
	default:
		return fmt.Errorf("unsupported macOS package format: %s", ext)
	}
}

// installDmg mounts and installs from DMG
func (i *Installer) installDmg() error {
	// Mount DMG
	mountCmd := exec.Command("hdiutil", "attach", i.downloadPath, "-nobrowse")
	mountOut, err := mountCmd.Output()
	if err != nil {
		i.output.WriteString(fmt.Sprintf("Failed to mount DMG: %v\n", err))
		return fmt.Errorf("failed to mount DMG: %w", err)
	}

	// Parse mount point from output
	mountPoint := strings.TrimSpace(string(mountOut))
	i.log.Info(fmt.Sprintf("DMG mounted at: %s", mountPoint))
	i.output.WriteString(fmt.Sprintf("DMG mounted at: %s\n", mountPoint))

	// Find .app in mounted volume
	appPath := filepath.Join(mountPoint, i.appName+".app")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		// Try to find any .app
		files, _ := filepath.Glob(filepath.Join(mountPoint, "*.app"))
		if len(files) > 0 {
			appPath = files[0]
		}
	}

	// Copy to Applications
	if _, err := os.Stat(appPath); err == nil {
		copyCmd := exec.Command("cp", "-R", appPath, "/Applications/")
		copyCmd.Stdout = &i.output
		copyCmd.Stderr = &i.output
		if err := copyCmd.Run(); err != nil {
			i.output.WriteString(fmt.Sprintf("Failed to copy to Applications: %v\n", err))
			return fmt.Errorf("failed to copy to Applications: %w", err)
		}
		i.log.Info("Application installed to /Applications/")
		i.output.WriteString("Application installed to /Applications/\n")
	}

	// Unmount DMG
	detachCmd := exec.Command("hdiutil", "detach", mountPoint)
	detachCmd.Stdout = &i.output
	detachCmd.Stderr = &i.output
	detachCmd.Run()

	return nil
}

// installZip extracts and installs from ZIP
func (i *Installer) installZip() error {
	// Extract to temp
	tmpDir := filepath.Join(os.TempDir(), "fyclip-update")
	os.MkdirAll(tmpDir, 0755)

	cmd := exec.Command("unzip", "-o", i.downloadPath, "-d", tmpDir)
	cmd.Stdout = &i.output
	cmd.Stderr = &i.output
	if err := cmd.Run(); err != nil {
		i.output.WriteString(fmt.Sprintf("Failed to extract: %v\n", err))
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Find .app
	appPath := filepath.Join(tmpDir, i.appName+".app")
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		files, _ := filepath.Glob(filepath.Join(tmpDir, "*.app"))
		if len(files) > 0 {
			appPath = files[0]
		}
	}

	// Copy to Applications
	if _, err := os.Stat(appPath); err == nil {
		copyCmd := exec.Command("cp", "-R", appPath, "/Applications/")
		copyCmd.Stdout = &i.output
		copyCmd.Stderr = &i.output
		if err := copyCmd.Run(); err != nil {
			i.output.WriteString(fmt.Sprintf("Failed to copy to Applications: %v\n", err))
			return fmt.Errorf("failed to copy to Applications: %w", err)
		}
		i.log.Info("Application installed to /Applications/")
		i.output.WriteString("Application installed to /Applications/\n")
	}

	// Cleanup
	os.RemoveAll(tmpDir)

	return nil
}

// AutoUpdater provides a complete update workflow
type AutoUpdater struct {
	checker     *Checker
	downloadDir string
	log         *logger.Logger
}

// NewAutoUpdater creates a new auto updater
func NewAutoUpdater(owner, repo, currentVersion string, opts ...Option) *AutoUpdater {
	checker := NewChecker(owner, repo, currentVersion, opts...)
	return &AutoUpdater{
		checker:     checker,
		downloadDir: os.TempDir(),
		log:         logger.Get(),
	}
}

// CheckAndDownload checks for updates and downloads if available
func (a *AutoUpdater) CheckAndDownload(ctx context.Context) (*UpdateInfo, *Downloader, error) {
	updateInfo, err := a.checker.CheckForUpdate(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Check if update is available
	if CompareVersions(updateInfo.LatestVersion, updateInfo.CurrentVersion) <= 0 {
		a.log.Info("No update available")
		return updateInfo, nil, nil
	}

	a.log.Info(fmt.Sprintf("Update available: %s -> %s", updateInfo.CurrentVersion, updateInfo.LatestVersion))

	// Download the update
	downloader := NewDownloader(a.checker, updateInfo)
	if err := downloader.Download(ctx, nil); err != nil {
		return nil, nil, err
	}

	return updateInfo, downloader, nil
}

// CheckAndInstall checks for updates, downloads and installs
func (a *AutoUpdater) CheckAndInstall(ctx context.Context, appName string) error {
	_, downloader, err := a.CheckAndDownload(ctx)
	if err != nil {
		return err
	}

	if downloader == nil {
		a.log.Info("Already on latest version")
		return nil
	}

	// Install the update
	installer := NewInstaller(downloader.GetDownloadPath(), appName)
	if err := installer.Install(); err != nil {
		return err
	}

	a.log.Info("Update installed successfully")
	return nil
}

// ParseRepoFromURL parses owner and repo from a GitHub URL
func ParseRepoFromURL(githubURL string) (owner, repo string, err error) {
	// Handle both HTTPS and SSH URLs
	// https://github.com/owner/repo
	// git@github.com:owner/repo.git

	githubURL = strings.TrimSuffix(githubURL, ".git")
	githubURL = strings.TrimPrefix(githubURL, "git@github.com:")
	githubURL = strings.TrimPrefix(githubURL, "https://github.com/")

	parts := strings.Split(githubURL, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub URL: %s", githubURL)
	}

	return parts[0], parts[1], nil
}

// ParseRepoFromModule parses owner and repo from go.mod module name
func ParseRepoFromModule(moduleName string) (owner, repo string, err error) {
	// Remove https:// or git@ prefix
	moduleName = strings.TrimPrefix(moduleName, "https://")
	moduleName = strings.TrimPrefix(moduleName, "git@github.com:")
	moduleName = strings.TrimSuffix(moduleName, ".git")

	parts := strings.Split(moduleName, "/")
	if len(parts) < 3 {
		// Could be just owner/repo
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
		return "", "", fmt.Errorf("invalid module name: %s", moduleName)
	}

	// Find the position where owner/repo starts
	// Usually it's the last two parts for github.com
	owner = parts[len(parts)-2]
	repo = parts[len(parts)-1]

	return owner, repo, nil
}

// GetAssetDownloadURL constructs a direct download URL for a specific release
func GetAssetDownloadURL(owner, repo, version, assetName string) string {
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, version, assetName)
}

// ValidateDownloadURL validates if a download URL is accessible
func ValidateDownloadURL(downloadURL string) error {
	resp, err := http.Head(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("URL returned status %d", resp.StatusCode)
	}

	return nil
}

// GetLatestReleaseTag gets the latest release tag from GitHub
func GetLatestReleaseTag(ctx context.Context, owner, repo string) (string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

// URLEncode encodes a string for use in URLs
func URLEncode(s string) string {
	return url.QueryEscape(s)
}
