// File: internal/update/checker_test.go
package update

import (
	"runtime"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		// Equal versions
		{"1.0.0", "1.0.0", 0},
		{"1.0", "1.0", 0},
		{"1", "1", 0},

		// v1 > v2
		{"2.0.0", "1.0.0", 1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.1", "1.0.0", 1},
		{"1.2.0", "1.1.9", 1},
		{"2.0", "1.9.9", 1},
		{"1.0.0", "0.9.9", 1},

		// v1 < v2
		{"1.0.0", "2.0.0", -1},
		{"1.0.0", "1.1.0", -1},
		{"1.0.0", "1.0.1", -1},
		{"1.1.9", "1.2.0", -1},
		{"1.9.9", "2.0", -1},

		// With v prefix
		{"v1.0.0", "v1.0.0", 0},
		{"v2.0.0", "v1.0.0", 1},
		{"v1.0.0", "v2.0.0", -1},

		// Different length
		{"1.0.0.1", "1.0.0", 1},
		{"1.0", "1.0.0", 0},
		{"1.0.0", "1.0", 0},

		// Pre-release handling: 1.0.0-beta < 1.0.0
		{"1.0.0-beta", "1.0.0", -1},
		{"1.0.0", "1.0.0-beta", 1},
		{"1.0.0-beta", "1.0.0-beta", 0},
		{"1.0.0-beta", "1.0.1-beta", -1},

		// Dev version handling: dev should be treated as equal to any version
		{"dev", "1.0.0", 0},
		{"1.0.0", "dev", 0},
		{"dev", "2.5.3", 0},
		{"dev", "dev", 0},
	}

	for _, tt := range tests {
		t.Run(tt.v1+"_"+tt.v2, func(t *testing.T) {
			result := CompareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestParseRepoFromURL(t *testing.T) {
	tests := []struct {
		url       string
		wantOwner string
		wantRepo  string
	}{
		{"https://github.com/Sarwarhridoy4/FyClip", "Sarwarhridoy4", "FyClip"},
		{"https://github.com/owner/repo", "owner", "repo"},
		{"git@github.com:owner/repo.git", "owner", "repo"},
		{"git@github.com:owner/repo", "owner", "repo"},
		{"owner/repo", "owner", "repo"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			owner, repo, err := ParseRepoFromURL(tt.url)
			if err != nil {
				t.Errorf("ParseRepoFromURL(%q) error = %v", tt.url, err)
				return
			}
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Errorf("ParseRepoFromURL(%q) = (%q, %q), want (%q, %q)",
					tt.url, owner, repo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}

func TestParseRepoFromModule(t *testing.T) {
	tests := []struct {
		module    string
		wantOwner string
		wantRepo  string
	}{
		{"github.com/Sarwarhridoy4/FyClip", "Sarwarhridoy4", "FyClip"},
		{"github.com/owner/repo", "owner", "repo"},
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			owner, repo, err := ParseRepoFromModule(tt.module)
			if err != nil {
				t.Errorf("ParseRepoFromModule(%q) error = %v", tt.module, err)
				return
			}
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Errorf("ParseRepoFromModule(%q) = (%q, %q), want (%q, %q)",
					tt.module, owner, repo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}

func TestFindAssetPrefersPlatformSpecificPackage(t *testing.T) {
	checker := NewChecker("owner", "repo", "1.0.0")

	assets := []GitHubAsset{
		{
			Name:        platformAssetName("portable.zip"),
			DownloadURL: "https://example.com/portable.zip",
			Size:        123,
			ContentType: "application/zip",
		},
		{
			Name:        platformAssetName(preferredAssetExtension()),
			DownloadURL: "https://example.com/preferred" + preferredAssetExtension(),
			Size:        456,
			ContentType: "application/octet-stream",
		},
	}

	asset := checker.findAsset(assets)
	if asset == nil {
		t.Fatal("expected matching asset")
	}

	if got, want := asset.Name, platformAssetName(preferredAssetExtension()); got != want {
		t.Fatalf("findAsset selected %q, want %q", got, want)
	}
}

func TestFindAssetSkipsAssetsMissingDownloadURL(t *testing.T) {
	checker := NewChecker("owner", "repo", "1.0.0")

	assets := []GitHubAsset{
		{
			Name:        platformAssetName(preferredAssetExtension()),
			DownloadURL: "",
		},
		{
			Name:        platformAssetName("fallback.zip"),
			DownloadURL: "https://example.com/fallback.zip",
		},
	}

	asset := checker.findAsset(assets)
	if asset == nil {
		t.Fatal("expected fallback asset")
	}

	if asset.DownloadURL == "" {
		t.Fatal("findAsset should not return asset without download URL")
	}
}

func TestFindAssetReturnsNilWithoutPlatformMatch(t *testing.T) {
	checker := NewChecker("owner", "repo", "1.0.0")

	assets := []GitHubAsset{
		{
			Name:        "fyclip-solaris-sparc.pkg",
			DownloadURL: "https://example.com/fyclip-solaris-sparc.pkg",
		},
	}

	if asset := checker.findAsset(assets); asset != nil {
		t.Fatalf("expected no asset match, got %q", asset.Name)
	}
}

func preferredAssetExtension() string {
	switch runtime.GOOS {
	case "linux":
		return ".deb"
	case "windows":
		return ".exe"
	case "darwin":
		return ".dmg"
	default:
		return ".zip"
	}
}

func platformAssetName(ext string) string {
	var osPart string
	switch runtime.GOOS {
	case "linux":
		osPart = "linux"
	case "windows":
		osPart = "windows"
	case "darwin":
		osPart = "darwin"
	default:
		osPart = runtime.GOOS
	}

	var archPart string
	switch runtime.GOARCH {
	case "amd64":
		archPart = "amd64"
	case "arm64":
		archPart = "arm64"
	case "386":
		archPart = "386"
	default:
		archPart = runtime.GOARCH
	}

	return "fyclip-" + osPart + "-" + archPart + ext
}
