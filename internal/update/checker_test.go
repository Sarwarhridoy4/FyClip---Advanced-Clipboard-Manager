// File: internal/update/checker_test.go
package update

import (
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
		url      string
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
		module   string
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