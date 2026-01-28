# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-01-28

### Added

#### AI-Powered Analysis
- **OpenAI Integration**: Optional AI analysis of flagged suspicious commits using GPT-4 Mini
- **Multi-step Reasoning**: Improved prompting with system and user prompts for better analysis
- **Structured Output**: AI analysis returns confidence scores and detailed reasoning
- **Token Efficient**: Only analyzes already-flagged commits (not all commits)
- **Cost Effective**: ~$0.00003 per commit analyzed using GPT-4 Mini

#### Remote Repository Support
- **GitHub URL Support**: Analyze repositories by passing GitHub URLs directly
- **URL Parsing**: Automatically parse GitHub blob/tree URLs and extract branch information
- **Auto-cloning**: Seamless cloning of remote repositories with automatic cleanup
- **Branch Fallback**: Graceful fallback to default branch if specified branch doesn't exist
- **Temporary Management**: Proper cleanup of temporary cloned directories

#### Configuration Improvements
- **Auto-detection**: Automatically loads `cadence.yml` from current directory if present
- **AI Config**: New AI section in configuration for enabling/disabling AI analysis
- **Environment Variables**: Support for `CADENCE_AI_KEY` environment variable
- **Sample Config**: Updated sample config includes AI configuration options

#### Skills.sh Integration
- **Skills Manifest**: Added `skills.json` defining Cadence as an agent skill
- **Skill Definitions**: Two main skills: `analyze-repository` and `detect-suspicious-commit`
- **Agent Compatible**: Compatible with Claude Code, GitHub Copilot, and other agent platforms
- **Standardized Output**: Skill output conforms to skills.sh specifications
- **Documentation**: Added `SKILLS.md` with integration examples and usage guide

#### Detection Enhancements
- **AI Detection Indicators**: System trained to identify:
  - Template-like code patterns
  - "Too perfect" code structure
  - Missing error handling
  - Pattern repetition
  - Lack of domain-specific optimizations

### Fixed
- Fixed unused variable `reportData` in text reporter tests
- Eliminated AI slop patterns detected in our own codebase
- Improved code clarity without breaking functionality

### Changed

- **Prompting Logic**: Enhanced with multi-step reasoning for better accuracy
- **Report Output**: Added optional AI analysis to both text and JSON reports
- **Error Handling**: Improved branch not found errors with helpful fallback
- **Variable Naming**: Improved generic variable names for better readability
  - Renamed `jobsData` to `jobList` in webhook handlers
  - Fixed inconsistent variable naming in test files
- **Comment Cleanup**: Removed verbose TODO comments and replaced with concise documentation
  - Streamlined webhook handler placeholder comments
  - Eliminated AI slop patterns identified by our own detection system
- **Code Consistency**: Applied consistent naming conventions across codebase
- **Error Handling**: Reviewed and maintained idiomatic Go error handling patterns

### Technical Details

- **New Dependencies**:
  - github.com/sashabaranov/go-openai - OpenAI API client
- **New Packages**:
  - `internal/ai` - AI analysis engine with OpenAI support
- **Updated Packages**:
  - `cmd/cadence/analyze.go` - Remote repository support
  - `internal/git/repository.go` - Branch fallback logic
  - `internal/config/config.go` - AI configuration
  - `internal/detector/detector.go` - AI analysis field
  - `internal/reporter/` - AI analysis output
- Used Cadence's own AI slop detection to identify and clean up problematic patterns
- Self-analysis confirmed reduced AI pattern indicators
- Maintained backward compatibility while improving code quality

---

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

[0.1.2]: https://github.com/CodeMeAPixel/Cadence/releases/tag/v0.1.2
[0.1.1]: https://github.com/CodeMeAPixel/Cadence/releases/tag/v0.1.1
[0.1.0]: https://github.com/CodeMeAPixel/Cadence/releases/tag/v0.1.0
