#!/usr/bin/env powershell
# Build Cadence for all platforms

# Get version info from git
$version = (git describe --tags 2>$null) -replace '-[0-9]*-g[0-9a-f]*$', ''
if (-not $version) { $version = "0.1.0" }
$commit = git rev-parse --short HEAD 2>$null
if (-not $commit) { $commit = "unknown" }
$buildTime = [DateTime]::UtcNow.ToString("yyyy-MM-ddTHH:mm:ssZ")

$ldflags = "-ldflags=-X github.com/trycadence/cadence/internal/version.Version=$version -X github.com/trycadence/cadence/internal/version.GitCommit=$commit -X github.com/trycadence/cadence/internal/version.BuildTime=$buildTime"

Write-Host "Building Cadence for all platforms..."
Write-Host "  Version: $version"
Write-Host "  Commit:  $commit"
Write-Host "  Time:    $buildTime"
Write-Host ""

# Create bin directory if it doesn't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# Build for Linux x64
Write-Host "Building for Linux x64..."
$env:GOOS = "linux"
$env:GOARCH = "amd64"
& go build $ldflags -o bin/cadence-linux-amd64 ./cmd/cadence
if ($LASTEXITCODE -ne 0) { exit 1 }

# Build for macOS x64
Write-Host "Building for macOS x64..."
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
& go build $ldflags -o bin/cadence-darwin-amd64 ./cmd/cadence
if ($LASTEXITCODE -ne 0) { exit 1 }

# Build for macOS ARM64
Write-Host "Building for macOS ARM64..."
$env:GOOS = "darwin"
$env:GOARCH = "arm64"
& go build $ldflags -o bin/cadence-darwin-arm64 ./cmd/cadence
if ($LASTEXITCODE -ne 0) { exit 1 }

# Build for Windows x64
Write-Host "Building for Windows x64..."
$env:GOOS = "windows"
$env:GOARCH = "amd64"
& go build $ldflags -o bin/cadence-windows-amd64.exe ./cmd/cadence
if ($LASTEXITCODE -ne 0) { exit 1 }

# Clean up environment
Remove-Item env:GOOS -ErrorAction SilentlyContinue
Remove-Item env:GOARCH -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "Built all platforms:"
Write-Host "  bin/cadence-linux-amd64"
Write-Host "  bin/cadence-darwin-amd64"
Write-Host "  bin/cadence-darwin-arm64"
Write-Host "  bin/cadence-windows-amd64.exe"
