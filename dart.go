package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// DartCLI wraps Dart command-line operations.
type DartCLI struct {
	workDir     string
	credentials *PubCredentials
	hostedURL   string
}

// NewDartCLI creates a new DartCLI instance.
func NewDartCLI(workDir string) *DartCLI {
	return &DartCLI{workDir: workDir}
}

// SetCredentials sets the pub credentials.
func (d *DartCLI) SetCredentials(creds *PubCredentials) {
	d.credentials = creds
}

// SetHostedURL sets a custom pub hosted URL.
func (d *DartCLI) SetHostedURL(url string) {
	d.hostedURL = url
}

// Analyze runs dart analyze.
func (d *DartCLI) Analyze(ctx context.Context) error {
	return d.run(ctx, "dart", "analyze", "--fatal-infos", "--fatal-warnings")
}

// FormatCheck checks code formatting.
func (d *DartCLI) FormatCheck(ctx context.Context) error {
	return d.run(ctx, "dart", "format", "--set-exit-if-changed", ".")
}

// Test runs package tests.
func (d *DartCLI) Test(ctx context.Context, cfg TestConfig) error {
	args := []string{"test"}
	if cfg.Platform != "" {
		args = append(args, "--platform", cfg.Platform)
	}
	if cfg.Concurrency > 0 {
		args = append(args, "--concurrency", strconv.Itoa(cfg.Concurrency))
	}
	if cfg.Coverage {
		args = append(args, "--coverage")
	}

	return d.run(ctx, "dart", args...)
}

// FlutterTest runs Flutter tests.
func (d *DartCLI) FlutterTest(ctx context.Context) error {
	return d.run(ctx, "flutter", "test")
}

// PublishDryRun runs publish dry-run validation.
func (d *DartCLI) PublishDryRun(ctx context.Context) error {
	return d.run(ctx, "dart", "pub", "publish", "--dry-run")
}

// Publish publishes the package.
func (d *DartCLI) Publish(ctx context.Context, force bool) error {
	args := []string{"pub", "publish"}
	if force {
		args = append(args, "--force")
	}

	cmd := exec.CommandContext(ctx, "dart", args...)
	cmd.Dir = d.workDir

	// Set credentials if available
	env := os.Environ()
	if d.credentials != nil && d.credentials.AccessToken != "" {
		env = append(env, fmt.Sprintf("PUB_TOKEN=%s", d.credentials.AccessToken))
	}
	if d.hostedURL != "" {
		env = append(env, fmt.Sprintf("PUB_HOSTED_URL=%s", d.hostedURL))
	}
	cmd.Env = env

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errOutput := strings.TrimSpace(stderr.String())
		if errOutput != "" {
			return fmt.Errorf("%s: %w", errOutput, err)
		}
		return err
	}

	return nil
}

// GetDependencies runs pub get.
func (d *DartCLI) GetDependencies(ctx context.Context) error {
	return d.run(ctx, "dart", "pub", "get")
}

// GetVersion returns the Dart version.
func (d *DartCLI) GetVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "dart", "--version")
	cmd.Dir = d.workDir

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Dart version: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// run executes a command.
func (d *DartCLI) run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = d.workDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errOutput := strings.TrimSpace(stderr.String())
		if errOutput != "" {
			return fmt.Errorf("%s: %w", errOutput, err)
		}
		return err
	}

	return nil
}
