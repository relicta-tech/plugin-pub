package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadCredentials(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid credentials",
			content: `{
				"accessToken": "test-token",
				"refreshToken": "refresh-token",
				"tokenEndpoint": "https://accounts.google.com/o/oauth2/token",
				"expiration": 9999999999
			}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			content: `{invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			path := filepath.Join(tempDir, "credentials.json")

			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write file: %v", err)
			}

			creds, err := LoadCredentials(path)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if creds.AccessToken != "test-token" {
				t.Errorf("expected access token 'test-token', got %s", creds.AccessToken)
			}
		})
	}
}

func TestLoadCredentials_FileNotFound(t *testing.T) {
	_, err := LoadCredentials("/nonexistent/credentials.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestCreateCredentialsFromToken(t *testing.T) {
	token := "my-access-token"
	creds := CreateCredentialsFromToken(token)

	if creds.AccessToken != token {
		t.Errorf("expected access token %s, got %s", token, creds.AccessToken)
	}
	if creds.RefreshToken != "" {
		t.Errorf("expected empty refresh token, got %s", creds.RefreshToken)
	}
}

func TestPubCredentials_IsExpired(t *testing.T) {
	tests := []struct {
		name       string
		expiration int64
		expected   bool
	}{
		{
			name:       "no expiration",
			expiration: 0,
			expected:   false,
		},
		{
			name:       "future expiration",
			expiration: time.Now().Unix() + 3600, // 1 hour from now
			expected:   false,
		},
		{
			name:       "past expiration",
			expiration: time.Now().Unix() - 3600, // 1 hour ago
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &PubCredentials{
				AccessToken: "token",
				Expiration:  tt.expiration,
			}

			if creds.IsExpired() != tt.expected {
				t.Errorf("expected IsExpired() = %v, got %v", tt.expected, creds.IsExpired())
			}
		})
	}
}

func TestPubCredentials_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		creds    *PubCredentials
		expected bool
	}{
		{
			name: "valid credentials",
			creds: &PubCredentials{
				AccessToken: "token",
				Expiration:  time.Now().Unix() + 3600,
			},
			expected: true,
		},
		{
			name: "empty token",
			creds: &PubCredentials{
				AccessToken: "",
				Expiration:  time.Now().Unix() + 3600,
			},
			expected: false,
		},
		{
			name: "expired credentials",
			creds: &PubCredentials{
				AccessToken: "token",
				Expiration:  time.Now().Unix() - 3600,
			},
			expected: false,
		},
		{
			name: "no expiration set",
			creds: &PubCredentials{
				AccessToken: "token",
				Expiration:  0,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.creds.IsValid() != tt.expected {
				t.Errorf("expected IsValid() = %v, got %v", tt.expected, tt.creds.IsValid())
			}
		})
	}
}
