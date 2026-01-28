# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-01-28

### Added

#### Core Analysis Engine
- Git repository analysis for detecting suspicious commit patterns
- Statistical anomaly detection using z-score analysis
- Commit pair analysis and velocity calculations
- Confidence scoring (0.0-1.0) based on triggered detection strategies
- Repository statistics including percentiles (P50, P75, P90, P95, P99)

#### Detection Strategies
- **VelocityStrategy**: Detects abnormally fast coding speeds
- **SizeStrategy**: Flags commits with suspiciously large additions/deletions
- **TimingStrategy**: Identifies rapid-fire commits (batch processing indicator)
- **MergeCommitStrategy**: Analyzes merge commit patterns
- **RatioStrategy**: Detects unusual addition/deletion ratios (>90% additions)
- **DisperseStrategy**: Identifies unusually dispersed file changes
- **PrecisionStrategy**: Detects pattern anomalies in commit statistics
- Pluggable strategy interface for custom detection logic

#### Reporting
- Text report format with confidence scores and detailed metrics
- JSON report format for programmatic integration
- Customizable output filtering and presentation
- Detailed reason explanations for each flagged commit

#### Webhook Server
- GitHub webhook support for real-time repository monitoring
- GitLab webhook support with token-based authentication
- HMAC-SHA256 signature verification for security
- Async job processing with configurable worker pool
- REST API for job management:
  - `POST /webhooks/github` - GitHub webhook endpoint
  - `POST /webhooks/gitlab` - GitLab webhook endpoint
  - `GET /jobs/:id` - Check job status and results
  - `GET /jobs?limit=N` - List recent analysis jobs
  - `GET /health` - Health check endpoint
- In-memory job queue (100 job capacity)

#### CLI Interface
- `cadence analyze` - Analyze a repository with customizable thresholds
- `cadence config` - Generate and manage configuration files
- `cadence webhook` - Start webhook server for repository integration
- `cadence version` - Display version and build information
- Command-line flags for all configuration options
- Support for YAML configuration files
- Environment variable support for all settings

#### Configuration
- YAML-based configuration with sensible defaults
- Per-threshold customization:
  - Suspicious additions/deletions limits
  - Velocity limits (additions/deletions per minute)
  - Minimum time delta between commits
  - File exclusion patterns (glob support)
- Environment variable overrides
- Command-line flag overrides

#### Error Handling
- Custom error types with context
- Type-safe error handling throughout codebase
- Helpful error messages for debugging

#### Build & Deployment
- Version information embedded at compile time (git tag, commit, build time)
- Cross-platform builds (Linux, macOS, Windows)
- Multi-architecture support (amd64, arm64)
- Makefile for common development tasks
- GitHub Actions CI/CD pipeline:
  - Automated testing (Go 1.21, 1.22, 1.23)
  - Code formatting checks
  - Linting with golangci-lint
  - Security scanning with Gosec
  - Coverage reporting (85% threshold)
  - Artifact generation for releases
- GitHub Actions Release workflow:
  - Automatic binary builds for all platforms
  - SHA256 checksums
  - Release notes generation

#### Testing
- 70+ test cases covering all packages
- Unit tests for detection strategies
- Integration tests for end-to-end workflows
- Webhook server tests with signature verification
- Configuration loading and validation tests
- Repository analysis tests
- Git operation tests
- Metrics and statistics tests
- Reporter output tests

#### Documentation
- README with quick start guide
- Usage examples and output examples
- Configuration guide with presets
- API endpoint documentation
- Webhook setup instructions
- Contributing guidelines for custom strategies

### Technical Details

- **Language**: Go 1.23.0
- **Dependencies**:
  - go-git/v5 - Git operations
  - spf13/cobra - CLI framework
  - spf13/viper - Configuration management
  - gofiber/fiber/v2 - HTTP web framework
  - google/uuid - Job ID generation
- **Architecture**:
  - Modular package design
  - Clean separation of concerns
  - Interface-based extensibility
  - Async job processing with worker pool
  - In-memory job storage

### Known Limitations

- Job queue limited to 100 in-memory jobs (oldest dropped when exceeded)
- Single-machine deployment only (no distributed queue support)
- Statistical analysis requires minimum commit history for accurate baselines
- File exclusion uses glob patterns (not regex)

---

[0.1.0]: https://github.com/CodeMeAPixel/Cadence/releases/tag/v0.1.0
