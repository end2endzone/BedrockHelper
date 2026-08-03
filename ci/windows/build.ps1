$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = "$PSScriptRoot/../.."

Set-Location $ProjectRoot/src

$Version = (Get-Content $ProjectRoot\VERSION).Trim()
$Commit = (git rev-parse --short HEAD).Trim()
$Date = (Get-Date).ToString("yyyy-MM-dd")

# Point this to the exact package path where your main function lives
$Pkg = "main"

Write-Host "Building BedrockHelper version $Version..."

# Run build from the root, pointing to the main package directory

go build -ldflags "-X '$Pkg.Version=$Version' -X '$Pkg.CommitHash=$Commit' -X '$Pkg.BuildDate=$Date'" -o $ProjectRoot/bin/bedrock_helper.exe ./cmd/bedrock_helper

Write-Host "Build complete!"
