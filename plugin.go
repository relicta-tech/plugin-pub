package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/relicta-tech/relicta-plugin-sdk/helpers"
	"github.com/relicta-tech/relicta-plugin-sdk/plugin"
)

// Version is set at build time.
var Version = "0.1.0"

// Config represents Pub plugin configuration.
type Config struct {
	PubspecPath     string     `json:"pubspec_path"`
	UpdateVersion   bool       `json:"update_version"`
	CredentialsPath string     `json:"credentials_path"`
	AccessToken     string     `json:"access_token"`
	HostedURL       string     `json:"hosted_url"`
	Validate        bool       `json:"validate"`
	Analyze         bool       `json:"analyze"`
	FormatCheck     bool       `json:"format_check"`
	Test            bool       `json:"test"`
	TestConfig      TestConfig `json:"test_config"`
	DryRunValidate  bool       `json:"dry_run_validate"`
	Force           bool       `json:"force"`
	Exclude         []string   `json:"exclude"`
	DryRun          bool       `json:"dry_run"`
}

// TestConfig defines test execution options.
type TestConfig struct {
	Platform    string `json:"platform"`
	Concurrency int    `json:"concurrency"`
	Coverage    bool   `json:"coverage"`
}

// PubPlugin implements the Dart/Flutter Pub plugin.
type PubPlugin struct{}

// GetInfo returns plugin metadata.
func (p *PubPlugin) GetInfo() plugin.Info {
	return plugin.Info{
		Name:        "pub",
		Version:     Version,
		Description: "Dart/Flutter package publishing to pub.dev",
		Hooks: []plugin.Hook{
			plugin.HookPrePublish,
			plugin.HookPostPublish,
		},
	}
}

// Validate validates plugin configuration.
func (p *PubPlugin) Validate(ctx context.Context, config map[string]any) (*plugin.ValidateResponse, error) {
	cfg := p.parseConfig(config)
	vb := helpers.NewValidationBuilder()

	// Check Dart installation
	if _, err := exec.LookPath("dart"); err != nil {
		vb.AddError("dart", "Dart SDK not found in PATH")
	}

	// Check pubspec.yaml
	pubspecPath := cfg.PubspecPath
	if pubspecPath == "" {
		pubspecPath = "pubspec.yaml"
	}

	pubspec, err := ParsePubspec(pubspecPath)
	if err != nil {
		vb.AddError("pubspec_path", fmt.Sprintf("Invalid pubspec.yaml: %v", err))
	} else {
		if err := ValidatePubspec(pubspec); err != nil {
			// Use AddError for pubspec validation issues as they are important
			vb.AddError("pubspec", err.Error())
		}
	}

	// Credentials are optional in validation - just check if they exist
	// The actual authentication will be validated at runtime

	return vb.Build(), nil
}

// Execute runs the plugin for a given hook.
func (p *PubPlugin) Execute(ctx context.Context, req plugin.ExecuteRequest) (*plugin.ExecuteResponse, error) {
	cfg := p.parseConfig(req.Config)
	cfg.DryRun = cfg.DryRun || req.DryRun
	logger := slog.Default().With("plugin", "pub", "hook", req.Hook)

	switch req.Hook {
	case plugin.HookPrePublish:
		return p.executePrePublish(ctx, &req.Context, cfg, logger)
	case plugin.HookPostPublish:
		return p.executePostPublish(ctx, &req.Context, cfg, logger)
	default:
		return &plugin.ExecuteResponse{
			Success: true,
			Message: fmt.Sprintf("Hook %s not handled by pub plugin", req.Hook),
		}, nil
	}
}

func (p *PubPlugin) executePrePublish(ctx context.Context, releaseCtx *plugin.ReleaseContext, cfg *Config, logger *slog.Logger) (*plugin.ExecuteResponse, error) {
	version := releaseCtx.Version
	logger = logger.With("version", version)

	pubspecPath := cfg.PubspecPath
	if pubspecPath == "" {
		pubspecPath = "pubspec.yaml"
	}

	// Parse pubspec to get package info
	pubspec, err := ParsePubspec(pubspecPath)
	if err != nil {
		return &plugin.ExecuteResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse pubspec.yaml: %v", err),
		}, nil
	}

	logger = logger.With("package", pubspec.Name)

	// Determine if this is a Flutter package
	isFlutter := IsFlutterPackage(pubspec)
	if isFlutter {
		logger = logger.With("flutter", true)
	}

	dart := NewDartCLI(".")

	// Update version in pubspec.yaml
	if cfg.UpdateVersion {
		logger.Info("Updating version in pubspec.yaml")
		if cfg.DryRun {
			logger.Info("[DRY-RUN] Would update version", "from", pubspec.Version, "to", version)
		} else {
			if err := UpdateVersion(pubspecPath, version); err != nil {
				return &plugin.ExecuteResponse{
					Success: false,
					Message: fmt.Sprintf("Failed to update version: %v", err),
				}, nil
			}
		}
	}

	// Run dart analyze
	if cfg.Analyze {
		logger.Info("Running dart analyze")
		if cfg.DryRun {
			logger.Info("[DRY-RUN] Would run dart analyze")
		} else {
			if err := dart.Analyze(ctx); err != nil {
				return &plugin.ExecuteResponse{
					Success: false,
					Message: fmt.Sprintf("Analysis failed: %v", err),
				}, nil
			}
		}
	}

	// Run format check
	if cfg.FormatCheck {
		logger.Info("Checking code formatting")
		if cfg.DryRun {
			logger.Info("[DRY-RUN] Would check code formatting")
		} else {
			if err := dart.FormatCheck(ctx); err != nil {
				return &plugin.ExecuteResponse{
					Success: false,
					Message: fmt.Sprintf("Format check failed: %v", err),
				}, nil
			}
		}
	}

	// Run tests
	if cfg.Test {
		logger.Info("Running tests")
		if cfg.DryRun {
			logger.Info("[DRY-RUN] Would run tests", "config", cfg.TestConfig, "flutter", isFlutter)
		} else {
			if isFlutter {
				if err := dart.FlutterTest(ctx); err != nil {
					return &plugin.ExecuteResponse{
						Success: false,
						Message: fmt.Sprintf("Flutter tests failed: %v", err),
					}, nil
				}
			} else {
				if err := dart.Test(ctx, cfg.TestConfig); err != nil {
					return &plugin.ExecuteResponse{
						Success: false,
						Message: fmt.Sprintf("Tests failed: %v", err),
					}, nil
				}
			}
		}
	}

	// Run dry-run validation
	if cfg.DryRunValidate {
		logger.Info("Running publish dry-run validation")
		if cfg.DryRun {
			logger.Info("[DRY-RUN] Would run dart pub publish --dry-run")
		} else {
			if err := dart.PublishDryRun(ctx); err != nil {
				return &plugin.ExecuteResponse{
					Success: false,
					Message: fmt.Sprintf("Dry-run validation failed: %v", err),
				}, nil
			}
		}
	}

	logger.Info("PrePublish completed successfully")
	return &plugin.ExecuteResponse{
		Success: true,
		Message: "Package validated successfully",
	}, nil
}

func (p *PubPlugin) executePostPublish(ctx context.Context, releaseCtx *plugin.ReleaseContext, cfg *Config, logger *slog.Logger) (*plugin.ExecuteResponse, error) {
	version := releaseCtx.Version
	logger = logger.With("version", version)

	pubspecPath := cfg.PubspecPath
	if pubspecPath == "" {
		pubspecPath = "pubspec.yaml"
	}

	// Parse pubspec to get package info
	pubspec, err := ParsePubspec(pubspecPath)
	if err != nil {
		return &plugin.ExecuteResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to parse pubspec.yaml: %v", err),
		}, nil
	}

	logger = logger.With("package", pubspec.Name)

	// Setup credentials
	var creds *PubCredentials
	if cfg.AccessToken != "" {
		creds = CreateCredentialsFromToken(cfg.AccessToken)
	} else {
		creds, _ = LoadCredentials(cfg.CredentialsPath)
	}

	dart := NewDartCLI(".")
	dart.SetCredentials(creds)

	// Set hosted URL if configured
	if cfg.HostedURL != "" {
		dart.SetHostedURL(cfg.HostedURL)
	}

	// Publish
	logger.Info("Publishing to pub.dev")
	if cfg.DryRun {
		logger.Info("[DRY-RUN] Would publish package",
			"package", pubspec.Name,
			"version", version,
			"force", cfg.Force)
	} else {
		if err := dart.Publish(ctx, cfg.Force); err != nil {
			return &plugin.ExecuteResponse{
				Success: false,
				Message: fmt.Sprintf("Publish failed: %v", err),
			}, nil
		}
	}

	var msg string
	if cfg.DryRun {
		msg = fmt.Sprintf("[DRY-RUN] Would publish %s@%s to pub.dev", pubspec.Name, version)
	} else {
		msg = fmt.Sprintf("Published %s@%s to pub.dev", pubspec.Name, version)
	}

	logger.Info("PostPublish completed successfully")
	return &plugin.ExecuteResponse{
		Success: true,
		Message: msg,
	}, nil
}

func (p *PubPlugin) parseConfig(raw map[string]any) *Config {
	parser := helpers.NewConfigParser(raw)

	// Parse test config
	testConfig := TestConfig{
		Platform:    "vm",
		Concurrency: 4,
		Coverage:    false,
	}
	if testRaw, ok := raw["test_config"].(map[string]any); ok {
		if platform, ok := testRaw["platform"].(string); ok {
			testConfig.Platform = platform
		}
		if conc, ok := testRaw["concurrency"].(float64); ok {
			testConfig.Concurrency = int(conc)
		}
		if cov, ok := testRaw["coverage"].(bool); ok {
			testConfig.Coverage = cov
		}
	}

	// Parse exclude list
	var exclude []string
	if excludeRaw, ok := raw["exclude"].([]any); ok {
		for _, e := range excludeRaw {
			if s, ok := e.(string); ok {
				exclude = append(exclude, s)
			}
		}
	}

	return &Config{
		PubspecPath:     parser.GetString("pubspec_path", "", "pubspec.yaml"),
		UpdateVersion:   parser.GetBool("update_version", true),
		CredentialsPath: parser.GetString("credentials_path", "", ""),
		AccessToken:     parser.GetString("access_token", "PUB_ACCESS_TOKEN", ""),
		HostedURL:       parser.GetString("hosted_url", "PUB_HOSTED_URL", ""),
		Validate:        parser.GetBool("validate", true),
		Analyze:         parser.GetBool("analyze", true),
		FormatCheck:     parser.GetBool("format_check", true),
		Test:            parser.GetBool("test", true),
		TestConfig:      testConfig,
		DryRunValidate:  parser.GetBool("dry_run_validate", true),
		Force:           parser.GetBool("force", true),
		Exclude:         exclude,
		DryRun:          parser.GetBool("dry_run", false),
	}
}
