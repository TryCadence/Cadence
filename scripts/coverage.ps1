# Cadence Code Coverage Script (Windows)
# Generates coverage reports and enforces coverage thresholds

param(
    [int]$Threshold = 85
)

$ErrorActionPreference = "Stop"

$CoverageFile = "coverage.out"
$CoverageHtml = "coverage.html"

Write-Host "Running tests with coverage..." -ForegroundColor Cyan

# Run tests with coverage
go test -v -coverprofile=$CoverageFile -covermode=atomic ./...

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Tests failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "📊 Generating HTML coverage report..." -ForegroundColor Cyan
go tool cover -html=$CoverageFile -o $CoverageHtml

Write-Host ""
Write-Host "📈 Coverage Summary:" -ForegroundColor Cyan
$summary = go tool cover -func=$CoverageFile | Select-Object -Last 1
Write-Host $summary

Write-Host ""

# Calculate coverage percentage
$coverageOutput = go tool cover -func=$CoverageFile | Select-String "total"
$coverage = [regex]::Match($coverageOutput, '(\d+\.\d+)%').Groups[1].Value

Write-Host "Total Coverage: ${coverage}%"
Write-Host "Threshold: ${Threshold}%"
Write-Host ""

# Check against threshold
if ([double]$coverage -lt [double]$Threshold) {
    Write-Host "❌ Coverage ${coverage}% is below ${Threshold}% threshold" -ForegroundColor Red
    Write-Host "📁 Open $CoverageHtml to view coverage details"
    exit 1
}

Write-Host "✅ Coverage ${coverage}% meets or exceeds ${Threshold}% threshold" -ForegroundColor Green
Write-Host ""
Write-Host "📁 Coverage report: $CoverageHtml"
Write-Host "💡 Run 'go tool cover -html=$CoverageFile' to view details in browser"
