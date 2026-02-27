# Cadence

Detects AI-generated content in **git repositories** and **websites**.

Cadence analyzes suspicious commits via code patterns, velocity anomalies, and statistical markers. It can also scan websites for AI-generated text using 38 detection strategies across 7 categories, with optional AI validation via OpenAI or Anthropic.

**Status**: Ready to use | **Tests**: 17 packages passing | **Go**: 1.24  
**Full Documentation**: [noslop.tech](https://noslop.tech) | [Quick Start](https://noslop.tech/docs/quick-start) | [CLI Commands](https://noslop.tech/docs/cli-commands)

## Features

- **38 Detection Strategies** across velocity, structural, behavioral, statistical, pattern, linguistic, and accessibility categories
- **Git Repository Analysis** — commit velocity, size, timing, naming patterns, burst detection, error handling patterns, and more
- **Website Content Analysis** — overused phrases, generic language, excessive structure, AI vocabulary, accessibility issues
- **Real-Time Streaming** — SSE endpoints for live analysis progress and detection events
- **Multi-Provider AI** — Optional GPT-4o Mini or Claude analysis of flagged items
- **5 Report Formats** — JSON, Text, HTML, YAML, BSON
- **Plugin System** — Register custom detection strategies at runtime
- **Extensible Architecture** — Source-agnostic pipeline: `AnalysisSource` → `Detector` → `AnalysisReport`

## Quick Start

```bash
git clone https://github.com/TryCadence/Cadence.git
cd Cadence
make build
```

The Makefile auto-detects your OS (PowerShell on Windows, shell on Unix/Linux/macOS).

```bash
# Generate default config
./cadence config > cadence.yaml

# Analyze a git repository
./cadence analyze /path/to/repo -o report.json

# Analyze a website
./cadence web https://example.com --json --output report.json
```

**Alternative build methods:**

```bash
./scripts/build.sh          # Linux/macOS
.\scripts\build.ps1         # Windows
.\scripts\build-all.ps1     # Cross-compile all platforms
go build ./cmd/cadence       # Direct Go (no version injection)
```

## Usage

### Analyze a Repository

```bash
# Auto-loads cadence.yaml from current directory
./cadence analyze /path/to/repo -o report.txt

# With custom thresholds (overrides config)
./cadence analyze /path/to/repo \
  -o report.json \
  --suspicious-additions 500 \
  --max-additions-pm 100

# Analyze specific branch
./cadence analyze /path/to/repo -o report.json --branch main

# Exclude files
./cadence analyze /path/to/repo \
  -o report.json \
  --exclude-files "*.min.js,package-lock.json"
```

### Analyze Website Content

```bash
# Detect AI-generated content
./cadence web https://example.com

# JSON report to file
./cadence web https://example.com --json --output report.json

# With AI expert analysis (requires CADENCE_AI_KEY)
./cadence web https://example.com --config cadence.yaml --verbose
```

Web analysis detects:
- **Overused phrases** — "in today's world", "furthermore", "in conclusion"
- **Generic language** — "provide value", "stakeholder", "utilize"
- **AI vocabulary** — characteristic word choices and phrasing
- **Excessive structure** — too many bullet points, perfect formatting
- **Repetitive patterns** — sentences starting with same words
- **Accessibility issues** — missing alt text, improper heading hierarchy

### Streaming API (SSE)

Real-time analysis via Server-Sent Events:

```bash
# Stream repository analysis
curl -N -X POST http://localhost:8000/api/stream/repository \
  -H "Content-Type: application/json" \
  -d '{"repository_url": "https://github.com/user/repo"}'

# Stream website analysis
curl -N -X POST http://localhost:8000/api/stream/website \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

SSE events: `progress` (phase updates), `detection` (each finding), `result` (final report), `error`.

## Detection Strategies

Cadence uses 38 strategies organized into 7 categories:

### Git Strategies (18)

| Strategy | Category | What It Detects |
|----------|----------|-----------------|
| Velocity Analysis | velocity | Code speed exceeding human norms (>100 additions/min) |
| Size Analysis | structural | Unusually large commits (>500 additions) |
| Timing Analysis | behavioral | Rapid-fire commits (<60 sec apart) |
| Burst Pattern | behavioral | Rapid-fire commit clusters suggesting batch processing |
| Ratio Analysis | statistical | Skewed addition/deletion ratios (>90% additions) |
| Precision Analysis | statistical | Suspiciously balanced code changes |
| Structural Consistency | statistical | Unnaturally consistent addition/deletion ratios |
| File Dispersion | structural | Too many files changed in a single commit |
| File Extension | structural | Suspicious bulk file creation patterns |
| Merge Commit Filter | structural | Unusual merge behavior and history rewrites |
| Commit Message | behavioral | AI-typical commit message patterns and phrasing |
| Naming Pattern | pattern | Generic or AI-typical variable/function naming |
| Error Handling | pattern | Missing or excessive error handling |
| Template Pattern | pattern | Boilerplate/template code from AI generation |
| Statistical Anomaly | statistical | Deviations from repository baseline (trimmed z-scores) |
| Timing Anomaly | behavioral | Unusual timing patterns between commits |
| Emoji Pattern | pattern | Excessive emoji usage in commit messages |
| Special Character | pattern | Unusual special character patterns |

### Web Strategies (20)

| Strategy | Category | What It Detects |
|----------|----------|-----------------|
| Overused Phrases | linguistic | Common AI filler phrases |
| Generic Language | linguistic | Excessive generic business language |
| AI Vocabulary | linguistic | AI-characteristic word choices |
| Perfect Grammar | linguistic | Suspiciously uniform sentence lengths |
| Missing Nuance | linguistic | Excessive absolute terms |
| Excessive Transitions | linguistic | Overuse of transition words |
| Excessive Structure | structural | Over-structured content with excessive lists |
| Heading Hierarchy | structural | Improper heading level order |
| Boilerplate Text | pattern | Common filler and boilerplate phrases |
| Repetitive Patterns | pattern | Repetitive sentence structures |
| Hardcoded Values | pattern | Inline styles, hardcoded pixels/colors |
| Generic Styling | pattern | Lack of CSS variables and theming |
| Emoji Overuse | pattern | Excessive emoji in content |
| Special Characters | pattern | Excessive special character patterns |
| Uniform Sentence Length | statistical | Unnaturally consistent sentence lengths |
| Missing Alt Text | accessibility | Images without alt text |
| Semantic HTML | accessibility | Overuse of divs instead of semantic tags |
| Accessibility Markers | accessibility | Missing ARIA labels and roles |
| Form Issues | accessibility | Missing labels, types, or names |
| Link Text Quality | accessibility | Generic or non-descriptive link text |

**Confidence-Weighted Scoring**: Each strategy has a confidence weight (0.0–1.0). Higher-confidence strategies contribute more to the overall score. Multiple signals compound.

## AI-Powered Analysis (Optional)

Cadence supports multiple AI providers for a second-opinion analysis of flagged items.

### Supported Providers

| Provider | Default Model | Setup |
|----------|---------------|-------|
| OpenAI | `gpt-4o-mini` | `CADENCE_AI_KEY=sk-...` |
| Anthropic | `claude-sonnet-4-20250514` | `CADENCE_AI_KEY=sk-ant-...` |

### Configuration

```yaml
# cadence.yaml
ai:
  enabled: true
  provider: "openai"     # or "anthropic"
  api_key: "sk-..."      # or use CADENCE_AI_KEY env var
  model: "gpt-4o-mini"
```

### AI Skills

Cadence includes 4 built-in AI skills:

| Skill | Purpose |
|-------|---------|
| `code_analysis` | Detect AI patterns in code snippets |
| `commit_review` | Holistic review of git commits |
| `pattern_explain` | Explain why a strategy flagged content |
| `report_summary` | Natural-language summary of analysis reports |

AI only analyzes already-flagged items — it never scans all commits.

## Report Formats

Cadence supports 5 output formats via the `AnalysisFormatter` interface:

| Format | Flag | Description |
|--------|------|-------------|
| Text | `-o report.txt` | Terminal-friendly with severity sections |
| JSON | `-o report.json` | Machine-readable with full metadata |
| HTML | `-o report.html` | Styled report with stat cards and charts |
| YAML | `-o report.yaml` | Config-friendly structured output |
| BSON | programmatic | Binary encoding for MongoDB integration |

All formats include: timing breakdown, source metrics, detection details, confidence scores, and assessment.

## Configuration

### Config File

```yaml
thresholds:
  suspicious_additions: 500
  suspicious_deletions: 1000
  max_additions_per_min: 100
  max_deletions_per_min: 500
  min_time_delta_seconds: 60

exclude_files:
  - "*.min.js"
  - "package-lock.json"
  - "yarn.lock"
  - "dist/**"
  - "build/**"
  - "vendor/**"
  - ".git/**"

# Disable specific detection strategies
strategies:
  disabled:
    - emoji_pattern_analysis
    - special_character_pattern_analysis

ai:
  enabled: true
  provider: "openai"
  model: "gpt-4o-mini"
```

`cadence.yaml` in the current directory is auto-loaded if no `--config` flag is specified.

### Command Line Flags

```bash
./cadence analyze <repo> [flags]

Flags:
  -o, --output string              Output file (required)
  --suspicious-additions int       Flag commits >N additions (default: 500)
  --suspicious-deletions int       Flag commits >N deletions (default: 1000)
  --max-additions-pm float         Max additions per minute (default: 100)
  --max-deletions-pm float         Max deletions per minute (default: 500)
  --min-time-delta int             Min seconds between commits (default: 60)
  --branch string                  Branch to analyze (default: all)
  --exclude-files strings          File patterns to exclude
  --config string                  Config file path
```

### Environment Variables

```bash
export CADENCE_AI_KEY="sk-..."              # AI provider API key
export CADENCE_WEBHOOK_PORT=3000            # Webhook server port
export CADENCE_WEBHOOK_SECRET="your-secret" # Webhook HMAC secret
export CADENCE_WEBHOOK_MAX_WORKERS=4        # Concurrent analysis workers
```

## Webhook Server & API

### Start the Server

```bash
./cadence webhook --port 8000 --secret "webhook-secret-key"
```

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/webhooks/github` | Receive GitHub push events |
| `POST` | `/webhooks/gitlab` | Receive GitLab push events |
| `POST` | `/api/stream/repository` | SSE streaming repository analysis |
| `POST` | `/api/stream/website` | SSE streaming website analysis |
| `GET` | `/jobs/:id` | Check job status |
| `GET` | `/jobs?limit=50` | List recent jobs |
| `GET` | `/health` | Health check |

### GitHub Webhook Setup

1. Repository Settings → Webhooks → Add webhook
2. Payload URL: `https://your-server:8000/webhooks/github`
3. Content type: `application/json`
4. Secret: same value as `--secret` flag
5. Events: Select "Push events"

## Architecture

```
AnalysisSource → Detector → AnalysisReport → Formatter
     │               │              │
     ├─ Git           ├─ Git (18)    ├─ JSON
     ├─ Web           ├─ Web (20)    ├─ Text
     └─ (extensible)  ├─ Plugin      ├─ HTML
                      └─ (custom)    ├─ YAML
                                     └─ BSON
```

### Project Structure

```
cmd/cadence/                   CLI commands (analyze, web, webhook, config, version)
internal/
  analysis/                    Core analysis framework
    adapters/git/              Git repository adapter & pattern strategies
    adapters/web/              Web content adapter & pattern strategies
    detectors/                 GitDetector, WebDetector
    sources/                   GitRepositorySource, WebsiteSource
    cache.go                   In-memory analysis cache with TTL
    observability.go           Metrics collection & Prometheus export
    plugin.go                  Plugin system for custom strategies
    registry.go                Strategy registry (38 strategies)
    runner.go                  Synchronous detection runner
    stream.go                  Streaming runner (SSE support)
    report.go                  Report model & scoring
    strategy.go                Strategy categories & metadata
    source.go                  Core interfaces (AnalysisSource, Detector)
  ai/                          AI provider system
    providers/openai/          OpenAI provider
    providers/anthropic/       Anthropic provider
    skills/                    Built-in AI skills (4)
    prompts/                   Prompt templates & response parsing
  config/                      Configuration loading & validation
  errors/                      Typed error system (CadenceError)
  logging/                     Structured logging (slog wrapper)
  metrics/                     Statistics & velocity calculations
  reporter/                    Report formatter factory
    formats/                   JSON, Text, HTML, YAML, BSON formatters
  version/                     Build version info
  webhook/                     Webhook server, handlers, SSE streaming
```

## Extending Cadence

### Plugin System

Register custom detection strategies at runtime:

```go
import "github.com/TryCadence/Cadence/internal/analysis"

type MyPlugin struct{}

func (p *MyPlugin) Info() analysis.StrategyInfo {
    return analysis.StrategyInfo{
        Name:        "my_custom_check",
        Category:    analysis.CategoryPattern,
        Confidence:  0.7,
        Description: "Detects my custom pattern",
        SourceTypes: []string{"git"},
    }
}

func (p *MyPlugin) Detect(ctx context.Context, data *analysis.SourceData) ([]analysis.Detection, error) {
    // Your detection logic
    return detections, nil
}

// Register with the plugin manager
pm := analysis.NewPluginManager()
pm.Register(&MyPlugin{})

// Use as a detector in the pipeline
runner := analysis.NewStreamingRunner()
events := runner.RunStream(ctx, source, pm.Detector())
```

### Custom Analysis Sources

Implement `AnalysisSource` to add new data sources:

```go
type NPMPackageSource struct {
    PackageName string
}

func (s *NPMPackageSource) Type() string { return "npm" }
func (s *NPMPackageSource) Validate(ctx context.Context) error { ... }
func (s *NPMPackageSource) Fetch(ctx context.Context) (*analysis.SourceData, error) { ... }
```

### Strategy Configuration

Disable strategies per-project via config:

```yaml
strategies:
  disabled:
    - emoji_pattern_analysis
    - special_character_pattern_analysis
    - missing_alt_text
```

## Development

```bash
make build     # Build with version injection
make test      # Run all tests
make cover     # Tests with coverage
make fmt       # Format code
make lint      # Run linter
make vet       # Run go vet
make clean     # Clean build artifacts
make help      # Show all targets
```

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Build for all platforms
.\scripts\build-all.ps1
```

## FAQ

**Can I use this in CI/CD?**  
Yes. Run `cadence analyze` in your pipeline, parse the JSON output, and fail the build if suspicious commits are found.

**How accurate is it?**  
Start with defaults and tune. Confidence-weighted scoring means multiple signals compound — a single low-confidence trigger rarely produces a false positive.

**What about legitimate fast commits?**  
The 30-second velocity floor prevents timestamp artifacts from inflating scores. Trimmed statistics (excluding top/bottom 10%) keep baselines robust.

**Does it work with GitHub/GitLab Enterprise?**  
Yes. Webhooks work with any Git host. Self-hosted instances need network access to your Cadence server.

**Which AI provider should I use?**  
Both work. OpenAI (`gpt-4o-mini`) is cheapest. Anthropic (`claude-sonnet-4-20250514`) may produce more nuanced analysis. The provider is selected in your config.

## Resources

- **Documentation**: [noslop.tech](https://noslop.tech)
- **Troubleshooting**: [Troubleshooting Guide](https://noslop.tech/docs/troubleshooting-guide)
- **Security**: [Security Documentation](https://noslop.tech/docs/security)
- **Contributing**: [Contributing Guide](https://noslop.tech/docs/contributing)
- **GitHub**: [TryCadence/Cadence](https://github.com/TryCadence/Cadence)

---

**Made with ❤️ by [CodeMeAPixel](https://codemeapixel.dev)**
