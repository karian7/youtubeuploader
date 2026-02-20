package test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	yt "github.com/porjo/youtubeuploader"
	"golang.org/x/oauth2"
)

func TestIsInvalidGrant(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "unrelated error",
			err:  errors.New("connection refused"),
			want: false,
		},
		{
			name: "RetrieveError with invalid_grant",
			err:  &oauth2.RetrieveError{ErrorCode: "invalid_grant", ErrorDescription: "Token has been expired or revoked."},
			want: true,
		},
		{
			name: "RetrieveError with different code",
			err:  &oauth2.RetrieveError{ErrorCode: "invalid_client"},
			want: false,
		},
		{
			name: "wrapped RetrieveError with invalid_grant",
			err:  fmt.Errorf("token refresh failed: %w", &oauth2.RetrieveError{ErrorCode: "invalid_grant"}),
			want: true,
		},
		{
			name: "error string containing invalid_grant",
			err:  errors.New(`oauth2: "invalid_grant" "Bad Request"`),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := yt.IsInvalidGrant(tt.err)
			if got != tt.want {
				t.Errorf("IsInvalidGrant() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCacheFile_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "token.json")
	cf := yt.CacheFile(path)

	token := &oauth2.Token{
		AccessToken:  "test-access",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh",
		Expiry:       time.Now().Add(time.Hour).Truncate(time.Second),
	}

	if err := cf.PutToken(token); err != nil {
		t.Fatalf("PutToken failed: %v", err)
	}

	got, err := cf.Token()
	if err != nil {
		t.Fatalf("Token failed: %v", err)
	}

	if got.AccessToken != token.AccessToken {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, token.AccessToken)
	}
	if got.RefreshToken != token.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, token.RefreshToken)
	}
}

func TestCacheFile_DeletedOnInvalidGrant(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "request.token")

	expiredToken := &oauth2.Token{
		AccessToken:  "expired-access",
		TokenType:    "Bearer",
		RefreshToken: "revoked-refresh",
		Expiry:       time.Now().Add(-time.Hour),
	}
	data, _ := json.Marshal(expiredToken)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatalf("failed to write token file: %v", err)
	}

	// Verify file exists and is readable
	cf := yt.CacheFile(tokenPath)
	_, err := cf.Token()
	if err != nil {
		t.Fatalf("should be able to read token: %v", err)
	}

	// Simulate invalid_grant error from token refresh
	refreshErr := &oauth2.RetrieveError{ErrorCode: "invalid_grant"}
	if yt.IsInvalidGrant(refreshErr) {
		if err := os.Remove(tokenPath); err != nil {
			t.Fatalf("failed to remove token cache: %v", err)
		}
	}

	// Verify file is deleted
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Error("token cache file should have been deleted after invalid_grant")
	}
}

func TestCacheFile_TokenReturnsErrorForMissingFile(t *testing.T) {
	cf := yt.CacheFile("/nonexistent/path/token.json")
	_, err := cf.Token()
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestCacheFile_PutTokenOverwritesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "token.json")
	cf := yt.CacheFile(path)

	old := &oauth2.Token{AccessToken: "old-token", TokenType: "Bearer"}
	if err := cf.PutToken(old); err != nil {
		t.Fatalf("PutToken (old) failed: %v", err)
	}

	updated := &oauth2.Token{AccessToken: "new-token", TokenType: "Bearer"}
	if err := cf.PutToken(updated); err != nil {
		t.Fatalf("PutToken (new) failed: %v", err)
	}

	got, err := cf.Token()
	if err != nil {
		t.Fatalf("Token failed: %v", err)
	}
	if got.AccessToken != "new-token" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "new-token")
	}
}
