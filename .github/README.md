# Cadence

Detects AI-generated content in **git repositories** and **websites**.

Analyze suspicious commits via code patterns, velocity anomalies, and statistical markers. Scan websites for AI-generated text using pattern detection and optional OpenAI validation.

**Status**: Ready to use | **Tests**: 70+ passing | **Go**: 1.23.0  
**📖 Full Documentation**: [noslop.tech](https://noslop.tech) | [Quick Start](https://noslop.tech/docs/quick-start) | [CLI Commands](https://noslop.tech/docs/cli-commands)

## Quick Start

For detailed setup instructions, see [Installation Guide](https://noslop.tech/docs/installation).

### Install

Version information (version, commit hash, build time) is automatically injected during build from git tags.

**Quick Start (all platforms)**

```bash
git clone https://github.com/codemeapixel/cadence.git
cd cadence

make build
```

The Makefile automatically detects your OS and uses the appropriate build method (PowerShell on Windows, shell on Unix/Linux/macOS).

**Alternative Methods**

```bash
# Using scripts directly
./scripts/build.sh        # Linux/macOS
.\scripts\build.ps1       # Windows

# Direct Go (no version injection)
go build ./cmd/cadence
```

Version injection is automatic when using `make build` or the platform-specific scripts.





### Analyze a Repository

```bash
# Generate default config
./cadence config > cadence.yaml

# Scan a repo for AI-generated commits
./cadence analyze /path/to/repo -o report.txt --config cadence.yaml
```

Output shows commits with unusual patterns, confidence scores, and reasons why each was flagged.

## Usage

For in-depth guides and examples, see:
- [Git Analysis Guide](https://noslop.tech/docs/git-analysis-guide)
- [Web Analysis Guide](https://noslop.tech/docs/web-analysis-guide)
- [CLI Commands Reference](https://noslop.tech/docs/cli-commands)

### Analyze a Repository

```bash
# Generate default config
./cadence config > cadence.yml

# Scan a repo (auto-loads cadence.yml if in current directory)
./cadence analyze /path/to/repo -o report.txt

# Or specify config explicitly
./cadence analyze /path/to/repo -o report.txt --config cadence.yml

# With custom thresholds (overrides config)
./cadence analyze /path/to/repo \
  -o report.json \
  --suspicious-additions 500 \
  --max-additions-pm 100

# Analyze specific branch
./cadence analyze /path/to/repo \
  -o report.json \
  --branch main

# Exclude certain files (node_modules, lock files, etc)
./cadence analyze /path/to/repo \
  -o report.json \
  --exclude-files "*.min.js,package-lock.json"
```

### Analyze Website Content for AI-Generated Text

```bash
# Detect AI-generated content on a website
./cadence web https://example.com

# Generate JSON report and save to file
./cadence web https://example.com --json --output report.json

# With AI expert analysis (requires CADENCE_AI_KEY)
./cadence web https://example.com --config cadence.yml --verbose
```

The `cadence web` command analyzes website content for common AI patterns:
- **Overused phrases**: "in today's world", "furthermore", "in conclusion"
- **Generic language**: "provide value", "various", "stakeholder", "utilize"
- **Excessive structure**: Too many bullet points, numbered lists, perfect formatting
- **Perfect grammar**: No contractions, no colloquialisms, suspiciously polished
- **Boilerplate text**: "our mission", "award-winning", "industry-leading"
- **Repetitive patterns**: Sentences starting with same words
- **Lack of nuance**: Few specific examples, no citations, vague references
- **Over-explanation**: Excessive transition phrases, explains obvious concepts

**Output Options**:
- Text format (default) - Human-readable with detailed pattern breakdown
- JSON format (`--json`) - Machine-readable with full metadata
- File output (`--output <file>`) - Save report to file instead of stdout

**Report Features**:
- Confidence score (0-100%) with assessment
- Detailed pattern breakdown with severity ratings
- Specific examples of flagged content (up to 5 per pattern)
- Context showing where patterns appear in the content
- Content quality metrics (word count, headings, quality score)



**Note**: `cadence.yml` in the current directory is automatically loaded if no `--config` flag is specified.

### Output Example (Text)

```
SUSPICIOUS COMMITS
Found 1 suspicious commit(s):

[1] Commit: a1b2c3d4
    Author:     John Doe <john@example.com>
    Date:       2024-01-27T10:30:00Z
    Confidence: 66.7%
    Additions:  1500 lines / 2000 total
    Deletions:  1200 lines / 1500 total
    Files:      45 files changed
    Time Delta: 0.50 minutes
    Velocity:   3000 additions/min | 2400 deletions/min
    
    Reasons:
    - Large commit: 1500 additions (threshold: 500)
    - Fast velocity: 3000 additions/min (threshold: 100)
```

### Output Example (JSON)

```json
{
  "suspicious_commits": [
    {
      "hash": "a1b2c3d4...",
      "author": "John Doe",
      "timestamp": "2024-01-27T10:30:00Z",
      "confidence_score": 0.667,
      "additions_filtered": 1500,
      "deletions_filtered": 1200,
      "addition_velocity_per_min": 3000.0,
      "reasons": [
        "Large commit: 1500 additions (threshold: 500)",
        "Fast velocity: 3000 additions/min (threshold: 100)"
      ]
    }
  ]
}
```

## Detection Strategies

For detailed strategy explanations, see [Detection Strategies Guide](https://noslop.tech/docs/detection-strategies).

Cadence flags commits that are suspicious based on:

| Strategy | What it looks for | Indicator |
|----------|------------------|-----------|
| **Velocity** | Abnormally fast coding | >100 additions/min |
| **Size** | Huge commits | >500 additions |
| **Timing** | Rapid-fire commits | <60 sec apart |
| **Additions Only** | No deletions, all adds | >90% additions |
| **Merge Pattern** | Unusual merge behavior | Context-dependent |

**Confidence Score**: Increases with each triggered strategy. Multiple signals = higher confidence.

## AI-Powered Analysis (Optional)

Cadence can leverage OpenAI's GPT models to analyze flagged commits for additional AI-generation indicators. This is **optional** and requires an OpenAI API key.

### Why Use AI Analysis?

- **Second opinion**: AI provides independent assessment of suspicious commits
- **Token efficient**: Only analyzes already-flagged commits (not all commits)
- **Lightweight**: Uses GPT-4 Mini for cost efficiency
- **Complementary**: Works alongside statistical detection, not instead of it

### Setup

1. Get an OpenAI API key from https://platform.openai.com/api-keys
2. Enable in config or environment:

```bash
# Via config file (cadence.yaml)
ai:
  enabled: true
  provider: "openai"
  api_key: "sk-..."  # or use env var below
  model: "gpt-4-mini"

# OR via environment variable
export CADENCE_AI_KEY="sk-..."
```

3. Run analysis as normal - AI kicks in automatically for suspicious commits

### Output

AI analysis appears in both text and JSON reports:

**Text Report:**
```
    AI Analysis:     likely AI-generated
```

**JSON Report:**
```json
"ai_analysis": "likely AI-generated"
```

### Cost Estimation

- Average suspicious commit: ~200 tokens
- GPT-4 Mini: ~$0.00015 per 1K tokens  
- Cost per analysis: ~$0.00003 (3 cents per 1000 commits)

## Configuration

For advanced configuration options, see [Configuration Guide](https://noslop.tech/docs/configuration).

### Config File (YAML)

Create a `cadence.yaml`:

```yaml
thresholds:
  # Commit size limits
  suspicious_additions: 500      # additions per commit
  suspicious_deletions: 1000     # deletions per commit
  
  # Velocity limits
  max_additions_per_min: 100     # additions per minute
  max_deletions_per_min: 500     # deletions per minute
  
  # Timing
  min_time_delta_seconds: 60     # seconds between commits

# Files to ignore
exclude_files:
  - "*.min.js"
  - "package-lock.json"
  - "yarn.lock"
```

### Command Line Flags

```bash
./cadence analyze <repo> [flags]

Flags:
  -o, --output string              Output file (required) - .txt or .json
  --suspicious-additions int       Flag commits >N additions (default: 500)
  --suspicious-deletions int       Flag commits >N deletions (default: 1000)
  --max-additions-pm float         Max additions per minute (default: 100)
  --max-deletions-pm float         Max deletions per minute (default: 500)
  --min-time-delta int            Min seconds between commits (default: 60)
  --branch string                 Branch to analyze (default: all)
  --exclude-files strings         File patterns to exclude
  --config string                 Config file path
```

### Environment Variables

```bash
# Set webhook server config
export CADENCE_WEBHOOK_PORT=3000
export CADENCE_WEBHOOK_SECRET="your-secret-key"
export CADENCE_WEBHOOK_MAX_WORKERS=4
```

## Webhook Server

For webhook setup and integration details, see [API Webhooks Guide](https://noslop.tech/docs/api-webhooks).

### Start the Server

```bash
./cadence webhook --port 3000 --secret "webhook-secret-key"
```

### Configure GitHub Webhook

1. Repository Settings → Webhooks → Add webhook
2. Payload URL: `https://your-server:3000/webhooks/github`
3. Content type: `application/json`
4. Secret: Use same value as `--secret` flag
5. Events: Select "Push events"

### API Endpoints

#### Receive webhook push event
```
POST /webhooks/github
POST /webhooks/gitlab
```

Returns:
```json
{
  "job_id": "uuid",
  "status": "processing"
}
```

#### Check job status
```
GET /jobs/:id
```

Returns:
```json
{
  "id": "job-uuid",
  "status": "completed|processing|pending|failed",
  "repo": "repo-name",
  "branch": "main",
  "timestamp": "2024-01-27T10:30:00Z",
  "result": {
    "suspicious_commits": [...]
  }
}
```

#### List recent jobs
```
GET /jobs?limit=50
```

#### Health check
```
GET /health
```

### How It Works

1. GitHub sends push webhook → HTTP POST to `/webhooks/github`
2. Cadence returns immediately with a job ID
3. Analysis happens in background (non-blocking)
4. Poll `/jobs/:id` to check progress
5. Results available when `status` is `completed`

## Common Questions

**Q: Can I use this in CI/CD?**  
A: Yes. Run `cadence analyze` in your pipeline, parse the JSON output, and fail the build if suspicious commits found.

**Q: How accurate is it?**  
A: Depends on your thresholds. Aggressive settings catch more but have more false positives. Start with defaults and tune.

**Q: What about non-AI code that looks suspicious?**  
A: The confidence score helps - legitimate fast commits might trigger one strategy but not multiple. Check the reasons.

**Q: Does it work with GitHub/GitLab Enterprise?**  
A: Webhooks work with any Git host. Self-hosted instances need network access to your Cadence server.

**Q: Can I extend it?**  
A: Yes. Detection strategies are pluggable interfaces in `internal/detector/`. Add custom logic easily.

## Development

For development setup and contribution guidelines, see [Build & Development Guide](https://noslop.tech/docs/build-development).

### Build

```bash
# Using Makefile (Linux/macOS)
make build

# Or direct Go (all platforms)
go build ./cmd/cadence

# Version info is automatically injected from git tags via go:generate
```

### Available Make Targets

```bash
make build    # Build binary with version injection
make install  # Install to $GOPATH/bin
make test     # Run all tests
make cover    # Run tests with coverage
make fmt      # Format code
make tidy     # Tidy dependencies  
make lint     # Run linter
make vet      # Run go vet
make run      # Run application
make clean    # Clean build artifacts
make help     # Show all targets
```

### Run Tests

```bash
go test ./...
go test -cover ./...  # With coverage
```

### Project Structure

```
cmd/cadence/          - CLI commands (analyze, webhook, config)
internal/
  analyzer/           - Repository analyzer orchestrator
  detector/           - Detection strategies
  git/                - Git operations
  metrics/            - Statistics and velocity calculations
  reporter/           - Output formatting (text, JSON)
  config/             - Configuration loading
  webhook/            - Webhook server (GitHub, GitLab)
  web/                - Website content fetching and analysis
    patterns/         - Web pattern detection strategies
  errors/             - Error types
test/                 - Integration tests
```

### Adding Custom Detection Strategies

**For Git Commit Analysis:**

Create a new strategy in `internal/detector/`:

```go
type CustomStrategy struct{}

func (s *CustomStrategy) Name() string {
    return "custom_detection"
}

func (s *CustomStrategy) Detect(pair *git.CommitPair, stats *metrics.RepositoryStats) (bool, string) {
    if isCustomSuspicious(pair) {
        return true, "Your reason here"
    }
    return false, ""
}
```

Register it in `internal/detector/detector.go` and it will automatically be used.

**For Web Content Analysis:**

Create a custom pattern strategy:

```go
import "github.com/codemeapixel/cadence/internal/web/patterns"

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

Or implement the `WebPatternStrategy` interface for more complex logic:

```go
type MyCustomStrategy struct{}

func (s *MyCustomStrategy) Name() string {
    return "my_custom_pattern"
}

func (s *MyCustomStrategy) Detect(content string, wordCount int) *patterns.DetectionResult {
    // Your detection logic here
    if detected {
        return &patterns.DetectionResult{
            Detected:    true,
            Type:        s.Name(),
            Severity:    0.8,
            Description: "Custom pattern detected",
            Examples:    []string{"example1", "example2"},
        }
    }
    return nil
}
```

## Resources

- **Documentation**: [noslop.tech](https://noslop.tech)
- **Troubleshooting**: [Troubleshooting Guide](https://noslop.tech/docs/troubleshooting-guide)
- **Security**: [Security Documentation](https://noslop.tech/docs/security)
- **Contributing**: [Contributing Guide](https://noslop.tech/docs/contributing)
- **GitHub**: [CodeMeAPixel/Cadence](https://github.com/CodeMeAPixel/Cadence)

---

*Made with ❤️ by [CodeMeAPixel](https://codemeapixel.dev) | Contact: hey@codemeapixel.dev*