package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/relicta-tech/relicta-plugin-sdk/plugin"
)

func TestPubPlugin_GetInfo(t *testing.T) {
	p := &PubPlugin{}
	info := p.GetInfo()

	if info.Name != "pub" {
		t.Errorf("expected name 'pub', got %s", info.Name)
	}

	if info.Version != Version {
		t.Errorf("expected version %s, got %s", Version, info.Version)
	}

	if len(info.Hooks) != 2 {
		t.Errorf("expected 2 hooks, got %d", len(info.Hooks))
	}

	hooks := map[plugin.Hook]bool{
		plugin.HookPrePublish:  false,
		plugin.HookPostPublish: false,
	}
	for _, h := range info.Hooks {
		hooks[h] = true
	}

	for hook, found := range hooks {
		if !found {
			t.Errorf("expected hook %v not found", hook)
		}
	}
}

func TestPubPlugin_ParseConfig(t *testing.T) {
	p := &PubPlugin{}

	tests := []struct {
		name     string
		config   map[string]any
		expected *Config
	}{
		{
			name:   "default values",
			config: map[string]any{},
			expected: &Config{
				PubspecPath:    "pubspec.yaml",
				UpdateVersion:  true,
				Validate:       true,
				Analyze:        true,
				FormatCheck:    true,
				Test:           true,
				DryRunValidate: true,
				Force:          true,
				TestConfig: TestConfig{
					Platform:    "vm",
					Concurrency: 4,
					Coverage:    false,
				},
			},
		},
		{
			name: "custom values",
			config: map[string]any{
				"pubspec_path":     "packages/core/pubspec.yaml",
				"update_version":   false,
				"access_token":     "test-token",
				"hosted_url":       "https://private.pub.dev",
				"validate":         false,
				"analyze":          false,
				"format_check":     false,
				"test":             false,
				"dry_run_validate": false,
				"force":            false,
				"dry_run":          true,
			},
			expected: &Config{
				PubspecPath:    "packages/core/pubspec.yaml",
				UpdateVersion:  false,
				AccessToken:    "test-token",
				HostedURL:      "https://private.pub.dev",
				Validate:       false,
				Analyze:        false,
				FormatCheck:    false,
				Test:           false,
				DryRunValidate: false,
				Force:          false,
				DryRun:         true,
				TestConfig: TestConfig{
					Platform:    "vm",
					Concurrency: 4,
					Coverage:    false,
				},
			},
		},
		{
			name: "with test config",
			config: map[string]any{
				"test_config": map[string]any{
					"platform":    "chrome",
					"concurrency": float64(2),
					"coverage":    true,
				},
			},
			expected: &Config{
				PubspecPath:    "pubspec.yaml",
				UpdateVersion:  true,
				Validate:       true,
				Analyze:        true,
				FormatCheck:    true,
				Test:           true,
				DryRunValidate: true,
				Force:          true,
				TestConfig: TestConfig{
					Platform:    "chrome",
					Concurrency: 2,
					Coverage:    true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := p.parseConfig(tt.config)

			if cfg.PubspecPath != tt.expected.PubspecPath {
				t.Errorf("expected pubspec_path %s, got %s", tt.expected.PubspecPath, cfg.PubspecPath)
			}
			if cfg.UpdateVersion != tt.expected.UpdateVersion {
				t.Errorf("expected update_version %v, got %v", tt.expected.UpdateVersion, cfg.UpdateVersion)
			}
			if cfg.AccessToken != tt.expected.AccessToken {
				t.Errorf("expected access_token %s, got %s", tt.expected.AccessToken, cfg.AccessToken)
			}
			if cfg.HostedURL != tt.expected.HostedURL {
				t.Errorf("expected hosted_url %s, got %s", tt.expected.HostedURL, cfg.HostedURL)
			}
			if cfg.Validate != tt.expected.Validate {
				t.Errorf("expected validate %v, got %v", tt.expected.Validate, cfg.Validate)
			}
			if cfg.Analyze != tt.expected.Analyze {
				t.Errorf("expected analyze %v, got %v", tt.expected.Analyze, cfg.Analyze)
			}
			if cfg.FormatCheck != tt.expected.FormatCheck {
				t.Errorf("expected format_check %v, got %v", tt.expected.FormatCheck, cfg.FormatCheck)
			}
			if cfg.Test != tt.expected.Test {
				t.Errorf("expected test %v, got %v", tt.expected.Test, cfg.Test)
			}
			if cfg.DryRunValidate != tt.expected.DryRunValidate {
				t.Errorf("expected dry_run_validate %v, got %v", tt.expected.DryRunValidate, cfg.DryRunValidate)
			}
			if cfg.Force != tt.expected.Force {
				t.Errorf("expected force %v, got %v", tt.expected.Force, cfg.Force)
			}
			if cfg.DryRun != tt.expected.DryRun {
				t.Errorf("expected dry_run %v, got %v", tt.expected.DryRun, cfg.DryRun)
			}
			if cfg.TestConfig.Platform != tt.expected.TestConfig.Platform {
				t.Errorf("expected test platform %s, got %s", tt.expected.TestConfig.Platform, cfg.TestConfig.Platform)
			}
			if cfg.TestConfig.Concurrency != tt.expected.TestConfig.Concurrency {
				t.Errorf("expected test concurrency %d, got %d", tt.expected.TestConfig.Concurrency, cfg.TestConfig.Concurrency)
			}
			if cfg.TestConfig.Coverage != tt.expected.TestConfig.Coverage {
				t.Errorf("expected test coverage %v, got %v", tt.expected.TestConfig.Coverage, cfg.TestConfig.Coverage)
			}
		})
	}
}

func TestPubPlugin_Validate(t *testing.T) {
	p := &PubPlugin{}

	// Create a temporary pubspec.yaml for testing
	tempDir := t.TempDir()
	pubspecPath := filepath.Join(tempDir, "pubspec.yaml")
	validPubspec := `name: test_package
version: 1.0.0
description: A test package for testing the pub plugin implementation with sufficient length
environment:
  sdk: '>=3.0.0 <4.0.0'
`
	if err := os.WriteFile(pubspecPath, []byte(validPubspec), 0644); err != nil {
		t.Fatalf("failed to create pubspec: %v", err)
	}

	tests := []struct {
		name       string
		config     map[string]any
		wantErrors bool
		errorField string
	}{
		{
			name: "valid config",
			config: map[string]any{
				"pubspec_path": pubspecPath,
			},
			wantErrors: false,
		},
		{
			name: "missing pubspec",
			config: map[string]any{
				"pubspec_path": "/nonexistent/pubspec.yaml",
			},
			wantErrors: true,
			errorField: "pubspec_path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := p.Validate(context.Background(), tt.config)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Note: Dart CLI validation is skipped if Dart isn't installed
			hasExpectedError := false
			for _, e := range resp.Errors {
				if e.Field == tt.errorField {
					hasExpectedError = true
					break
				}
			}

			if tt.wantErrors && tt.errorField != "" && !hasExpectedError {
				// Check if the error might be about Dart not being installed
				for _, e := range resp.Errors {
					if e.Field == "dart" {
						// Dart not installed, skip field-specific validation
						return
					}
				}
				if len(resp.Errors) == 0 {
					t.Errorf("expected error for field %s, got none", tt.errorField)
				}
			}
		})
	}
}

func TestPubPlugin_Execute_DryRun(t *testing.T) {
	p := &PubPlugin{}

	// Create temp directory with pubspec.yaml
	tempDir := t.TempDir()
	pubspecPath := filepath.Join(tempDir, "pubspec.yaml")
	pubspec := `name: test_package
version: 1.0.0
description: A test package for testing the pub plugin implementation with sufficient length
environment:
  sdk: '>=3.0.0 <4.0.0'
`
	if err := os.WriteFile(pubspecPath, []byte(pubspec), 0644); err != nil {
		t.Fatalf("failed to create pubspec: %v", err)
	}

	config := map[string]any{
		"pubspec_path":     pubspecPath,
		"access_token":     "test-token",
		"dry_run":          true,
		"update_version":   false,
		"analyze":          false,
		"format_check":     false,
		"test":             false,
		"dry_run_validate": false,
	}

	releaseCtx := plugin.ReleaseContext{
		Version: "2.0.0",
	}

	// Test PrePublish
	req := plugin.ExecuteRequest{
		Hook:    plugin.HookPrePublish,
		Context: releaseCtx,
		Config:  config,
		DryRun:  true,
	}
	resp, err := p.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("PrePublish failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("PrePublish should succeed in dry-run mode: %s", resp.Message)
	}

	// Test PostPublish
	req.Hook = plugin.HookPostPublish
	resp, err = p.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("PostPublish failed: %v", err)
	}
	if !resp.Success {
		t.Errorf("PostPublish should succeed in dry-run mode: %s", resp.Message)
	}
}
