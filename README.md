# Pub Plugin for Relicta

A Relicta plugin for publishing Dart and Flutter packages to pub.dev.

## Features

- Publish packages to pub.dev and custom registries
- Automatic version updates in pubspec.yaml
- Code analysis with `dart analyze`
- Format checking with `dart format`
- Test execution before publishing
- Dry-run validation
- Flutter package support
- OAuth credential management

## Installation

Download the pre-built binary from the [releases page](https://github.com/relicta-tech/plugin-pub/releases) or build from source:

```bash
go build -o pub .
```

## Configuration

Add the plugin to your `relicta.yaml`:

```yaml
plugins:
  - name: pub
    enabled: true
    hooks:
      - PrePublish
      - PostPublish
    config:
      # Path to pubspec.yaml
      pubspec_path: "pubspec.yaml"

      # Update version in pubspec.yaml
      update_version: true

      # Credentials (use env var)
      access_token: ${PUB_ACCESS_TOKEN}

      # Or specify credentials file path
      # credentials_path: "~/.pub-cache/credentials.json"

      # Custom registry URL (for private packages)
      # hosted_url: "https://pub.my-company.com"

      # Pre-publish validation
      validate: true
      analyze: true
      format_check: true
      test: true

      # Test configuration
      test_config:
        platform: "vm"
        concurrency: 4
        coverage: false

      # Dry-run validation before publish
      dry_run_validate: true

      # Force publish (skip confirmations)
      force: true
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `PUB_ACCESS_TOKEN` | OAuth access token for pub.dev |
| `PUB_HOSTED_URL` | Custom registry URL (overrides config) |

## Authentication

### Using dart pub login

The recommended way to authenticate:

```bash
dart pub login
```

This stores credentials in `~/.pub-cache/credentials.json`.

### Using Access Token

For CI/CD environments, set the `PUB_ACCESS_TOKEN` environment variable:

```bash
export PUB_ACCESS_TOKEN=your-token-here
```

## Pubspec Requirements

Your pubspec.yaml must include:

```yaml
name: my_package
version: 1.0.0
description: >-
  A description between 60-180 characters for pub.dev listing.
  This helps users understand what your package does.

environment:
  sdk: '>=3.0.0 <4.0.0'
```

### Flutter Packages

For Flutter packages, also include:

```yaml
dependencies:
  flutter:
    sdk: flutter

flutter:
  uses-material-design: true
```

## Hooks

### PrePublish

Executed before the release is published:
- Updates version in pubspec.yaml
- Runs `dart analyze`
- Checks code formatting
- Runs tests (Dart or Flutter)
- Validates with `dart pub publish --dry-run`

### PostPublish

Executed after successful release:
- Publishes to pub.dev with `dart pub publish --force`
- Reports success/failure

## Dry Run

Test your configuration without publishing:

```bash
relicta publish --dry-run
```

## Flutter Support

The plugin automatically detects Flutter packages by checking for:
- `flutter` in dependencies
- `flutter` section in pubspec.yaml

For Flutter packages:
- Uses `flutter test` instead of `dart test`
- Includes Flutter-specific validations

## Troubleshooting

### Dart SDK not found

Ensure Dart is installed and in your PATH:

```bash
dart --version
```

### Credentials expired

Re-authenticate with:

```bash
dart pub login
```

### Analysis failed

Run analysis manually to see detailed errors:

```bash
dart analyze --fatal-infos --fatal-warnings
```

### Format check failed

Fix formatting issues:

```bash
dart format .
```

### Description too short

pub.dev requires descriptions between 60-180 characters. Update your pubspec.yaml description.

## Private Registries

To publish to a private registry:

```yaml
config:
  hosted_url: "https://pub.my-company.com"
  access_token: ${PRIVATE_PUB_TOKEN}
```

## Development

### Running tests

```bash
go test -v ./...
```

### Building

```bash
go build -o pub .
```

## License

MIT License - see [LICENSE](LICENSE) for details.
