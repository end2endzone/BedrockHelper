$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = "$PSScriptRoot/../.."

Set-Location $ProjectRoot/src

go test -v ./...
