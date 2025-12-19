package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePubspec(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		wantName    string
		wantVersion string
	}{
		{
			name: "valid pubspec",
			content: `name: my_package
version: 1.2.3
description: A sample package for testing purposes with enough characters to pass
environment:
  sdk: '>=3.0.0 <4.0.0'
`,
			wantErr:     false,
			wantName:    "my_package",
			wantVersion: "1.2.3",
		},
		{
			name: "flutter package",
			content: `name: my_flutter_app
version: 1.0.0+1
description: A Flutter application for testing with enough character count for validation
environment:
  sdk: '>=3.0.0 <4.0.0'
dependencies:
  flutter:
    sdk: flutter
flutter:
  uses-material-design: true
`,
			wantErr:     false,
			wantName:    "my_flutter_app",
			wantVersion: "1.0.0+1",
		},
		{
			name:    "invalid yaml",
			content: `name: [invalid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			path := filepath.Join(tempDir, "pubspec.yaml")

			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write file: %v", err)
			}

			pubspec, err := ParsePubspec(path)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if pubspec.Name != tt.wantName {
				t.Errorf("expected name %s, got %s", tt.wantName, pubspec.Name)
			}
			if pubspec.Version != tt.wantVersion {
				t.Errorf("expected version %s, got %s", tt.wantVersion, pubspec.Version)
			}
		})
	}
}

func TestUpdateVersion(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		newVersion string
		expected   string
		wantErr    bool
	}{
		{
			name: "simple version update",
			content: `name: my_package
version: 1.0.0
description: Test package
`,
			newVersion: "2.0.0",
			expected: `name: my_package
version: 2.0.0
description: Test package
`,
			wantErr: false,
		},
		{
			name: "flutter version with build number",
			content: `name: my_app
version: 1.0.0+1
description: Test app
`,
			newVersion: "2.0.0+5",
			expected: `name: my_app
version: 2.0.0+5
description: Test app
`,
			wantErr: false,
		},
		{
			name: "preserve comments",
			content: `name: my_package
# This is the package version
version: 1.0.0
description: Test package
`,
			newVersion: "3.0.0",
			expected: `name: my_package
# This is the package version
version: 3.0.0
description: Test package
`,
			wantErr: false,
		},
		{
			name: "no version field",
			content: `name: my_package
description: Test package
`,
			newVersion: "1.0.0",
			expected:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			path := filepath.Join(tempDir, "pubspec.yaml")

			if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write file: %v", err)
			}

			err := UpdateVersion(path, tt.newVersion)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			if string(result) != tt.expected {
				t.Errorf("expected:\n%s\n\ngot:\n%s", tt.expected, string(result))
			}
		})
	}
}

func TestValidatePubspec(t *testing.T) {
	tests := []struct {
		name    string
		pubspec *Pubspec
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid pubspec",
			pubspec: &Pubspec{
				Name:        "my_package",
				Version:     "1.0.0",
				Description: "A package description that is long enough to pass the validation requirement of 60 chars",
				Environment: map[string]string{"sdk": ">=3.0.0 <4.0.0"},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			pubspec: &Pubspec{
				Version:     "1.0.0",
				Description: "A package description that is long enough to pass the validation requirement of 60 chars",
				Environment: map[string]string{"sdk": ">=3.0.0 <4.0.0"},
			},
			wantErr: true,
			errMsg:  "package name is required",
		},
		{
			name: "missing version",
			pubspec: &Pubspec{
				Name:        "my_package",
				Description: "A package description that is long enough to pass the validation requirement of 60 chars",
				Environment: map[string]string{"sdk": ">=3.0.0 <4.0.0"},
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "missing description",
			pubspec: &Pubspec{
				Name:        "my_package",
				Version:     "1.0.0",
				Environment: map[string]string{"sdk": ">=3.0.0 <4.0.0"},
			},
			wantErr: true,
			errMsg:  "description is required for pub.dev",
		},
		{
			name: "description too short",
			pubspec: &Pubspec{
				Name:        "my_package",
				Version:     "1.0.0",
				Description: "Short description",
				Environment: map[string]string{"sdk": ">=3.0.0 <4.0.0"},
			},
			wantErr: true,
			errMsg:  "description should be at least 60 characters (currently 17)",
		},
		{
			name: "missing environment",
			pubspec: &Pubspec{
				Name:        "my_package",
				Version:     "1.0.0",
				Description: "A package description that is long enough to pass the validation requirement of 60 chars",
			},
			wantErr: true,
			errMsg:  "environment section is required",
		},
		{
			name: "missing SDK constraint",
			pubspec: &Pubspec{
				Name:        "my_package",
				Version:     "1.0.0",
				Description: "A package description that is long enough to pass the validation requirement of 60 chars",
				Environment: map[string]string{},
			},
			wantErr: true,
			errMsg:  "SDK constraint is required in environment section",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePubspec(tt.pubspec)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsFlutterPackage(t *testing.T) {
	tests := []struct {
		name     string
		pubspec  *Pubspec
		expected bool
	}{
		{
			name: "pure dart package",
			pubspec: &Pubspec{
				Name:         "dart_package",
				Dependencies: map[string]any{"http": "^1.0.0"},
			},
			expected: false,
		},
		{
			name: "flutter dependency",
			pubspec: &Pubspec{
				Name:         "flutter_package",
				Dependencies: map[string]any{"flutter": map[string]any{"sdk": "flutter"}},
			},
			expected: true,
		},
		{
			name: "flutter section",
			pubspec: &Pubspec{
				Name:    "flutter_app",
				Flutter: map[string]any{"uses-material-design": true},
			},
			expected: true,
		},
		{
			name: "both flutter dependency and section",
			pubspec: &Pubspec{
				Name:         "full_flutter",
				Dependencies: map[string]any{"flutter": map[string]any{"sdk": "flutter"}},
				Flutter:      map[string]any{"uses-material-design": true},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFlutterPackage(tt.pubspec)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
