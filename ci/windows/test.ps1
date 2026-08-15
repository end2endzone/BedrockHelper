$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = "$PSScriptRoot/../.."

Set-Location $ProjectRoot

# go test -v ./...
gotestsum --format github-actions --junitfile ..\test-report.xml
