$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = "$PSScriptRoot/../.."

Set-Location $ProjectRoot

if (Test-Path bin) { Remove-Item -Recurse -Force bin }
