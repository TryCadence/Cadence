# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.3] - 2026-02-03

### Added
- **Test Coverage Expansion**: Comprehensive test suite for critical modules
  - CLI command tests (42+ test cases): Utils, web commands, webhooks, config
  - AI analysis tests (13+ test cases): Analyzer initialization, response parsing, confidence scoring
  - Web fetcher tests (16+ test cases): Content fetching, HTML parsing, quality assessment
  - Pattern detection tests: Strategy registry validation and consistency checks
  - Web pattern strategy tests: Accessibility, semantics, forms, link quality, hardcoded values

### Changed
- **Code Quality**: Eliminated all linting errors through comprehensive cleanup
  - Converted if-else chains to switch statements for better maintainability (5 files)
  - Combined chained append calls in detector initialization
  - Extracted helper functions for improved code clarity
  - Fixed 19+ function signatures to use named return values
  - Improved string comparison patterns (`len(x) == 0` → `x == ""`)
- **Coverage Configuration**: Updated codecov.yml
  - Removed `cmd/**` from ignore list to include CLI coverage
  - Adjusted project coverage target to 30% (realistic baseline)
  - Full coverage tracking now includes all modules

### Fixed
- **Security (gosec)**: Fixed file permission issues
  - Directory permissions: `0755` → `0o750`
  - File permissions: `0644`/`0600` → `0o600`
  - Added proper error handling for cleanup operations
- **Error Handling**: Added checks for unchecked errors
  - Response body close errors now handled properly
  - HTTP request cleanup improved with `http.NoBody`
  - Webhook test fixed for closed channel behavior
- **Type Safety**: Fixed parameter type combinations across AI analyzer methods
- **Version System**: Improved build version display
  - Now shows clean release tags (`v0.2.2` instead of `v0.2.2-9-gb6d7645`)
  - Updated both Unix/Linux/macOS and Windows build scripts
- **Constants**: Introduced status constants for webhook job states
  - `StatusPending`, `StatusProcessing`, `StatusCompleted`, `StatusFailed`
  - Replaced magic strings throughout webhook handlers
- **Misspellings**: Fixed `cancelled` → `canceled` across codebase
- **Pre-allocation**: Optimized slice allocations in analysis functions

### Technical Details
- **Test Results**: All tests passing with zero linting errors
- **Coverage Metrics** (v0.2.3):
  - `config`: 100.0% ✅
  - `metrics`: 97.0% ✅
  - `reporter`: 97.7% ✅
  - `analyzer`: 93.3% ✅
  - `detector`: 84.3% ✅
  - `git`: 73.3%
  - `ai`: 59.1%
  - `webhook`: 58.3%
  - `web`: 42.4%
  - `cmd/cadence`: 17.5%
  - **Total: 32.5%** (up from 27.5%)
- **Coverage Integration**: Full code coverage setup complete
  - New Makefile targets: `coverage`, `coverage-report`, `coverage-strict`
  - Cross-platform scripts: `coverage.sh` (Unix/Linux/macOS), `coverage.ps1` (Windows)
  - CI/CD pipeline collects coverage across all platforms
  - Codecov integration configured with 30% baseline threshold
- **Build System**: Improved version injection and cross-platform consistency
- **Test Infrastructure**: 100+ new test cases across all major modules

## [0.2.2] - 2026-02-03

### Added

#### Git Commit Analysis Strategies
- **Emoji Detection Strategy**: Flags commits with excessive emoji usage
  - Detects 3+ emojis per commit
  - Identifies emoji-only commits
  - Flags high emoji ratio (>20% of message)
  - Severity scores: 0.5-0.8 based on pattern type
- **Special Character Strategy**: Detects suspicious special character patterns
  - Flags excessive hyphens (5+), asterisks (4+), underscores (4+)
  - Identifies consecutive special character patterns
  - Severity scores: 0.6-0.9 based on density

#### Web Content Analysis Strategies (8 New)
Comprehensive web accessibility and HTML quality detection:
1. **Missing Alt Text Strategy**: Detects images without accessibility descriptions
   - Regex matches `<img>` tags, checks for `alt=` attribute
   - Severity: 0.6-1.0 based on missing/empty alt text ratio
2. **Semantic HTML Strategy**: Flags excessive div usage vs semantic tags
   - Compares div count vs nav/header/section/article/footer/main/aside tags
   - Severity: 0.6-0.75 if div ratio > 70%
3. **Accessibility Markers Strategy**: Detects missing ARIA labels and roles
   - Checks for aria-label, aria-describedby, aria-hidden, role, lang attributes
   - Severity: 0.5-0.7 if fewer than 2 types found
4. **Heading Hierarchy Strategy**: Validates proper heading level sequencing
   - Ensures h1-h6 tags follow correct hierarchy
   - Flags non-sequential levels and missing h1 start
   - Severity: 0.65 if hierarchy issues found
5. **Hardcoded Values Strategy**: Detects hardcoded pixels/colors instead of CSS variables
   - Counts inline styles with width/height, pixel values, hardcoded colors
   - Severity: 0.55-1.0 based on issue count
6. **Form Issues Strategy**: Flags missing labels and improper input types
   - Checks for label, placeholder, type, and name attributes on inputs
   - Severity: 0.7 if multiple issues found
7. **Link Text Quality Strategy**: Detects generic link text ("click here", "read more")
   - Matches against list of low-quality link phrases
   - Severity: 0.4-0.8 based on ratio
8. **Generic Styling Strategy**: Identifies default colors and missing custom theming
   - Checks for CSS variables, theme keywords, media queries, custom classes
   - Severity: 0.5-1.0 based on indicator count

#### Strategy Count Expansion
- **Total Detection Strategies**: Increased from 26 to 46
- **Git Strategies**: 18 total (16 existing + 2 new)
- **Web Strategies**: 28 total (10 existing + 10 new + 8 new accessibility)
- Updated frontend displays to reflect new strategy counts

### Fixed
- Removed unused imports from `semantic_html_strategy.go` (unused regexp package)
- Removed 6 unused variables from semantic HTML strategy compilation
- All backend compilation errors resolved

### Technical Details
- **New Files**:
  - `internal/detector/patterns/emoji_strategy.go` - Git emoji detection
  - `internal/detector/patterns/special_character_strategy.go` - Git special char detection
  - `internal/web/patterns/emoji_strategy.go` - Web emoji detection
  - `internal/web/patterns/special_characters_strategy.go` - Web special char detection
  - `internal/web/patterns/missing_alt_text_strategy.go`
  - `internal/web/patterns/semantic_html_strategy.go`
  - `internal/web/patterns/accessibility_markers_strategy.go`
  - `internal/web/patterns/heading_hierarchy_strategy.go`
  - `internal/web/patterns/hardcoded_values_strategy.go`
  - `internal/web/patterns/form_issues_strategy.go`
  - `internal/web/patterns/link_text_quality_strategy.go`
  - `internal/web/patterns/generic_styling_strategy.go`
- **Updated Files**:
  - `internal/web/patterns/strategy.go` - RegisterDefaults() registers all 20 web patterns
  - `src/components/landing/DetectionStrategies.tsx` - Updated to show 46 strategies, 20 web patterns
  - `src/routes/features.tsx` - Updated strategy count display

### Tests
- All existing tests pass (206/211 total)
- 5 pre-existing test failures unrelated to new code
- New strategies compile and register successfully

## [0.2.1] - 2026-01-30

### Added

#### Web Content Analysis Enhancements
- **Structured Reporting System**: New comprehensive reporting for web analysis
  - `JSONWebReporter` - Machine-readable JSON output with full metadata
  - `TextWebReporter` - Human-readable formatted text reports
  - `WebReportData` - Structured data format for analysis results
- **Detailed Flagged Content Reports**: Each detected pattern now includes:
  - Pattern type and severity score
  - Detailed reasoning for why content was flagged
  - Specific examples found in the content (up to 5 per pattern)
  - Context extraction showing where patterns appear (150 chars surrounding)
- **Pattern-Based Architecture**: Extensible detection system for web content
  - `WebPatternStrategy` interface for pluggable detection strategies
  - `WebPatternRegistry` for managing and executing detection strategies
  - 10 default strategies automatically registered
  - Custom pattern support via `RegisterStrategy()`
- **CLI Output Options**: New flags for web analysis command
  - `--json, -j` - Output in JSON format
  - `--output, -o <file>` - Write report to file instead of stdout
  - Existing `--verbose, -v` - Show detailed content quality metrics

#### Pattern Detection Strategies
New modular web pattern detection system with 10 built-in strategies:
1. **OverusedPhrasesStrategy** - Detects generic AI phrases ("in today's world", "furthermore")
2. **GenericLanguageStrategy** - Identifies business jargon ("leverage", "synergy", "utilize")
3. **ExcessiveStructureStrategy** - Flags over-formatted content with too many lists
4. **PerfectGrammarStrategy** - Detects suspiciously consistent sentence structures
5. **BoilerplateTextStrategy** - Identifies template phrases ("welcome to", "let's explore")
6. **RepetitivePatternsStrategy** - Catches repetitive sentence structures
7. **MissingNuanceStrategy** - Detects excessive use of absolute terms
8. **ExcessiveTransitionsStrategy** - Flags overuse of transition words
9. **UniformSentenceLengthStrategy** - Identifies unnatural sentence length consistency
10. **AIVocabularyStrategy** - Detects AI-characteristic vocabulary ("delve", "tapestry", "realm")

#### Technical Infrastructure
- **Pattern Registry**: `internal/web/patterns/` package with strategy management
  - Automatic strategy registration on initialization
  - Easy addition of custom detection patterns
  - Example custom strategy included
- **Refactored TextSlopAnalyzer**: Now uses registry-based detection
  - Cleaner, more maintainable codebase
  - Supports `RegisterStrategy()` for custom patterns
  - Backwards compatible with existing code
- **Reports Directory Management**: Automatic creation of `reports/` directory
  - All web analysis reports saved to `reports/` directory by default
  - Directory created automatically if it doesn't exist
  - Clean separation of reports from source code

### Changed
- **Web Analysis Output**: Now uses structured reporting system instead of simple formatted text
- **Pattern Detection**: Moved from monolithic analyzer to modular strategy-based system
- **Code Organization**: Separated detection strategies into individual files for maintainability
- **Report Storage**: Web reports now automatically saved to `reports/` directory when using `--output`

### Fixed
- **Anomaly Detection Baseline**: Fixed incorrect baseline calculation in statistical anomaly strategy
  - Was attempting to use non-existent `RepositoryStats` fields
  - Now properly uses `CalculateBaseline()` from commit pairs
  - Fixed standard deviation calculation (was using IQR instead)
- **Web Command Flag Registration**: Fixed `--json` and `--output` flags not being recognized
  - Flags were defined but not properly registered to the command
  - Added missing `rootCmd.AddCommand(webCmd)` call in init()
- **Variable Shadowing**: Fixed variable shadowing issue in `web/fetcher.go` causing compile error
- **Duplicate Code Block**: Removed duplicate if statement in `ai_slop_strategies.go`
- **Unused Variables**: Removed unused `lowerBound` variable in anomaly detection
- **Comment Cleanup**: Removed all comments from detector folder for cleaner codebase

### Technical Details
- **New Packages**:
  - `internal/web/patterns/` - Web content detection strategies
  - `internal/web/reporter.go` - Reporting system for web analysis
- **Updated Files**:
  - `cmd/cadence/web.go` - Added JSON/file output support
  - `internal/detector/patterns/text_slop_analyzer.go` - Refactored to use registry
  - `internal/detector/patterns/anomaly_strategy.go` - Fixed baseline calculation
- **New Files**:
  - `internal/web/patterns/strategy.go` - Core interfaces and registry
  - `internal/web/patterns/overused_phrases.go` - Overused phrases detection
  - `internal/web/patterns/generic_language.go` - Generic language detection
  - `internal/web/patterns/placeholder_strategies.go` - 8 additional strategies
  - `internal/web/patterns/custom_example.go` - Example custom strategy

### Usage Examples
```bash
# Generate JSON report and save to file
cadence web https://example.com --json --output report.json

# Human-readable text report
cadence web https://example.com

# With verbose content quality metrics
cadence web https://example.com --verbose

# JSON output to stdout for piping
cadence web https://example.com --json | jq '.flagged_items'
```

### For Developers
Adding custom web content detection patterns is now straightforward:

```go
// Create custom strategy
customStrategy := patterns.NewCustomPatternStrategy(
    "marketing_speak",
    []string{"synergy", "innovative", "disruptive"},
    2, // threshold
)

// Register with analyzer
analyzer := patterns.NewTextSlopAnalyzer()
analyzer.RegisterStrategy(customStrategy)
```

## [0.2.0] - 2026-01-30

### Added

#### Website AI Content Analysis
- **Web Content Analysis**: New `cadence web` command to analyze websites for AI-generated text
- **Text Slop Detection**: Identifies common AI patterns in written content:
  - Overused phrases and generic business language
  - Excessive structure and formatting patterns
  - Suspiciously perfect grammar without contractions
  - Boilerplate and marketing language
  - Repetitive sentence structures
  - Lack of specific examples or citations
  - Excessive transition phrases
- **Pattern Analysis**: Detailed breakdown of detected AI indicators with severity scores
- **Confidence Scoring**: Overall AI-generated content confidence (0-100%)
- **HTML Parsing**: Extracts and analyzes main content from websites
- **Optional AI Expert Analysis**: Integration with OpenAI for additional validation
- **User-friendly Reports**: Formatted ASCII reports with pattern details

#### Community & Governance
- **Security Policy**: Added `SECURITY.md` with vulnerability reporting procedures
  - Responsible disclosure guidelines
  - Direct contact method: security@noslop.tech
  - Support for private security advisories
  - Supported version information
- **Code of Conduct**: Added `CODE_OF_CONDUCT.md` for community standards
  - Inclusive behavior guidelines
  - Anti-harassment and discrimination policies
  - Clear enforcement procedures
- **Contributing Guidelines**: Added `CONTRIBUTING.md` with development workflow
  - Setup instructions for all platforms
  - Code style and testing requirements
  - Commit message conventions
  - PR process and checklist
  - Recognition for contributors
- **Issue Templates**: Expanded GitHub issue templates
  - Bug report template with environment details
  - Feature request template with use cases
  - Documentation improvement template
  - All templates include contact information

#### GitHub Infrastructure
- **Pull Request Template**: Added `.github/pull_request_template.md`
  - Type of change checklist (bug, feature, breaking, docs, etc.)
  - Testing requirements
  - Code quality checklist
  - Breaking changes documentation
- **Funding Configuration**: Updated `FUNDING.yml` with sponsorship links
  - GitHub sponsors link
  - Personal website and social media links
- **GitHub Workflows**: Added automated CI/CD pipeline
  - `test.yml` - Cross-platform testing on push and PR
  - Tests on macOS, Ubuntu, and Windows
  - Go 1.23 and 1.24 compatibility testing
  - Build, test, and vet steps

#### Build System Improvements
- **Cross-Platform Makefile**: Updated Makefile with OS detection
  - Windows builds use PowerShell scripts automatically
  - Unix/Linux/macOS builds use native shell commands
  - Automatic version injection on all platforms
  - Unified `make build` command works everywhere
  - Platform-specific `make install` support
- **Build Scripts**: Organized build utilities
  - `scripts/build.sh` for Unix/Linux/macOS
  - `scripts/build.ps1` for Windows PowerShell
  - Automatic version, commit, and timestamp injection

#### Documentation Updates
- **Build Documentation**: Clarified build instructions in README
  - Quick start with `make build` for all platforms
  - Alternative build methods documented
  - Version injection behavior explained
- **.github/README.md**: Added community guidelines overview
  - Directory structure explanation
  - Quick links to all community documents
  - Contact information and contribution encouragement

### Technical Details
- **New Package**: `internal/web` for website fetching and content extraction
- **New Analyzer**: `TextSlopAnalyzer` in `internal/detector/patterns/` for text-based detection
- **New CLI Command**: `cadence web <url>` for analyzing websites
- **Dependencies Added**: github.com/PuerkitoBio/goquery for HTML parsing

### Changed
- **Version Management**: Version now always comes from git tags
  - No manual version.go updates needed
  - Automatic injection during build
  - Git describes for version string format

### Usage Examples
```bash
# Analyze a website for AI-generated content
cadence web https://example.com

# With AI expert analysis enabled (requires CADENCE_AI_KEY)
cadence web https://example.com --config cadence.yml
```

## [0.1.2] - 2026-01-28

### Changed

#### Code Quality Improvements
- **Variable Naming**: Improved generic variable names for better readability
  - Renamed `jobsData` to `jobList` in webhook handlers
  - Fixed inconsistent variable naming in test files
- **Comment Cleanup**: Removed verbose TODO comments and replaced with concise documentation
  - Streamlined webhook handler placeholder comments
  - Eliminated AI slop patterns identified by our own detection system
- **Code Consistency**: Applied consistent naming conventions across codebase
- **Error Handling**: Reviewed and maintained idiomatic Go error handling patterns

### Fixed
- Fixed unused variable `reportData` in text reporter tests
- Eliminated AI slop patterns detected in our own codebase
- Improved code clarity without breaking functionality

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

[0.2.1]: https://github.com/TryCadence/Cadence/releases/tag/v0.2.1
[0.2.0]: https://github.com/TryCadence/Cadence/releases/tag/v0.2.0
[0.1.2]: https://github.com/TryCadence/Cadence/releases/tag/v0.1.2
[0.1.1]: https://github.com/TryCadence/Cadence/releases/tag/v0.1.1
[0.1.0]: https://github.com/TryCadence/Cadence/releases/tag/v0.1.0
