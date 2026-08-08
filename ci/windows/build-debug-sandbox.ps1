$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = "$PSScriptRoot/../.."

Set-Location $ProjectRoot

if (Test-Path .debug-sandbox) { Remove-Item -Recurse -Force .debug-sandbox }
New-Item -ItemType Directory -Path .debug-sandbox | Out-Null
Copy-Item -Recurse src/testdata .debug-sandbox/testdata
Copy-Item src/testdata/addons/foobar.mcaddon ".debug-sandbox/testdata/servers/server_with_multiple_packs/foobar.mcaddon"

Write-Host "Sandbox ready!"
