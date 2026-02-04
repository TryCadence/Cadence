# Build script for Cadence - automatically injects version info from git tags
# Usage: .\scripts\build.ps1 [-Output cadence.exe] [-Install]

param(
    [string]$Output = "cadence.exe",
    [switch]$Install = $false
)

try {
    $rawVersion = & git describe --tags 2>$null
    if ($rawVersion) {
        # Remove -N-gHASH suffix to get just the tag
        $VERSION = $rawVersion -replace '-[0-9]+-g[0-9a-f]+$', ''
    } else {
        $VERSION = "0.1.0"
    }
} catch {
    $VERSION = "0.1.0"
}

# Get short commit hash
try {
    $COMMIT = & git rev-parse --short HEAD 2>$null
    if (-not $COMMIT) {
        $COMMIT = "unknown"
    }
} catch {
    $COMMIT = "unknown"
}

# Get build time in UTC
$BUILD_TIME = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

$LDFLAGS = "-ldflags=`"-X github.com/trycadence/cadence/internal/version.Version=$VERSION -X github.com/trycadence/cadence/internal/version.GitCommit=$COMMIT -X github.com/trycadence/cadence/internal/version.BuildTime=$BUILD_TIME`""

Write-Host "Building Cadence..." -ForegroundColor Cyan
Write-Host "  Version: $VERSION" -ForegroundColor Green
Write-Host "  Commit:  $COMMIT" -ForegroundColor Green
Write-Host "  Time:    $BUILD_TIME" -ForegroundColor Green
Write-Host ""

if ($Install) {
    Write-Host "Installing..." -ForegroundColor Cyan
    Invoke-Expression "go install $LDFLAGS ./cmd/cadence"
} else {
    Invoke-Expression "go build $LDFLAGS -o $Output ./cmd/cadence"
    Write-Host "Build complete: $Output" -ForegroundColor Green
}
