package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// PubCredentials represents pub.dev credentials.
type PubCredentials struct {
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"refreshToken"`
	TokenEndpoint string `json:"tokenEndpoint"`
	Expiration    int64  `json:"expiration"`
}

// LoadCredentials loads pub.dev credentials from a file.
func LoadCredentials(path string) (*PubCredentials, error) {
	if path == "" {
		// Default path
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, ".pub-cache", "credentials.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds PubCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

// CreateCredentialsFromToken creates credentials from an access token.
func CreateCredentialsFromToken(token string) *PubCredentials {
	return &PubCredentials{
		AccessToken: token,
	}
}

// IsExpired checks if the credentials are expired.
func (c *PubCredentials) IsExpired() bool {
	if c.Expiration == 0 {
		return false // No expiration set
	}
	return time.Now().Unix() > c.Expiration
}

// IsValid checks if credentials are valid (non-empty and not expired).
func (c *PubCredentials) IsValid() bool {
	if c.AccessToken == "" {
		return false
	}
	return !c.IsExpired()
}

// GetDefaultCredentialsPath returns the default credentials path.
func GetDefaultCredentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".pub-cache", "credentials.json"), nil
}
