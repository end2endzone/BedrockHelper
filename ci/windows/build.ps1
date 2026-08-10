$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = "$PSScriptRoot/../.."

Set-Location $ProjectRoot/src

$Version = (Get-Content $ProjectRoot\VERSION).Trim()
$Commit = (git rev-parse --short HEAD).Trim()
$Date = (Get-Date).ToString("yyyy-MM-dd")

# GOOS
if (-not [string]::IsNullOrEmpty($env:GOOS)) {
    Write-Output "GOOS is set to: $env:GOOS"
} else {
    $env:GOOS = (go env GOHOSTOS).Trim()
    Write-Output "GOOS is not set and is forced to: $env:GOOS"
}

# GOARCH
if (-not [string]::IsNullOrEmpty($env:GOARCH)) {
    Write-Output "GOARCH is set to: $env:GOARCH"
} else {
    # Detect OS type
    #$env:GOARCH = switch ((Get-CimInstance Win32_Processor).Architecture) {
    #    9       { "amd64" }
    #    12      { "arm64" }
    #    default { "unknown" }
    #}

    $env:GOARCH = (go env GOHOSTARCH).Trim()

    Write-Output "GOARCH is not set and is forced to: $env:GOARCH"
}

# Ensures your binary does not depend on host operating system C libraries, making the binary completely portable.
$env:CGO_ENABLED = "0"

# Point this to the exact package path where your main function lives
$Pkg = "main"

Write-Host "Building bedrock_helper-$env:GOOS-$env:GOARCH.exe version $Version..."

# Run build from the root, pointing to the main package directory
go build -ldflags "-X '$Pkg.Version=$Version' -X '$Pkg.CommitHash=$Commit' -X '$Pkg.BuildDate=$Date'" -o $ProjectRoot/bin/bedrock_helper-$env:GOOS-$env:GOARCH.exe ./cmd/bedrock_helper

Write-Host "Build complete!"
