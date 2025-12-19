package main

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

// Pubspec represents a parsed pubspec.yaml file.
type Pubspec struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Description  string            `yaml:"description"`
	Homepage     string            `yaml:"homepage,omitempty"`
	Repository   string            `yaml:"repository,omitempty"`
	Environment  map[string]string `yaml:"environment"`
	Dependencies map[string]any    `yaml:"dependencies"`
	DevDeps      map[string]any    `yaml:"dev_dependencies"`
	Flutter      map[string]any    `yaml:"flutter,omitempty"`
}

// ParsePubspec parses a pubspec.yaml file.
func ParsePubspec(path string) (*Pubspec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read pubspec.yaml: %w", err)
	}

	var pubspec Pubspec
	if err := yaml.Unmarshal(data, &pubspec); err != nil {
		return nil, fmt.Errorf("failed to parse pubspec.yaml: %w", err)
	}

	return &pubspec, nil
}

// UpdateVersion updates the version in pubspec.yaml.
func UpdateVersion(path, version string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read pubspec.yaml: %w", err)
	}

	// Use regex to preserve formatting and comments
	pattern := regexp.MustCompile(`(?m)^version:\s*.+$`)
	if !pattern.Match(data) {
		return fmt.Errorf("version field not found in pubspec.yaml")
	}

	newData := pattern.ReplaceAll(data, []byte(fmt.Sprintf("version: %s", version)))

	if err := os.WriteFile(path, newData, 0644); err != nil {
		return fmt.Errorf("failed to write pubspec.yaml: %w", err)
	}

	return nil
}

// ValidatePubspec validates pubspec.yaml contents for pub.dev requirements.
func ValidatePubspec(pubspec *Pubspec) error {
	if pubspec.Name == "" {
		return fmt.Errorf("package name is required")
	}

	if pubspec.Version == "" {
		return fmt.Errorf("version is required")
	}

	if pubspec.Description == "" {
		return fmt.Errorf("description is required for pub.dev")
	}

	// Pub.dev description requirements
	if len(pubspec.Description) < 60 {
		return fmt.Errorf("description should be at least 60 characters (currently %d)", len(pubspec.Description))
	}
	if len(pubspec.Description) > 180 {
		return fmt.Errorf("description should be at most 180 characters (currently %d)", len(pubspec.Description))
	}

	// Check SDK constraint
	if pubspec.Environment == nil {
		return fmt.Errorf("environment section is required")
	}

	sdk, ok := pubspec.Environment["sdk"]
	if !ok || sdk == "" {
		return fmt.Errorf("SDK constraint is required in environment section")
	}

	return nil
}

// IsFlutterPackage checks if the pubspec indicates a Flutter package.
func IsFlutterPackage(pubspec *Pubspec) bool {
	// Check if flutter is a dependency
	if _, hasFlutter := pubspec.Dependencies["flutter"]; hasFlutter {
		return true
	}

	// Check if there's a flutter section
	if len(pubspec.Flutter) > 0 {
		return true
	}

	return false
}

// GetPackageName returns the package name from pubspec.
func (p *Pubspec) GetPackageName() string {
	return p.Name
}

// GetVersion returns the version from pubspec.
func (p *Pubspec) GetVersion() string {
	return p.Version
}
