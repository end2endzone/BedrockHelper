$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = Convert-Path("$PSScriptRoot/../..")

Set-Location $ProjectRoot

if (Test-Path .debug-sandbox) { Remove-Item -Recurse -Force .debug-sandbox }
New-Item -ItemType Directory -Path .debug-sandbox | Out-Null
Copy-Item -Recurse testdata .debug-sandbox/testdata
Copy-Item testdata/addons/foobar.mcaddon ".debug-sandbox/testdata/servers/server_with_multiple_packs/foobar.mcaddon"

# Testing installing multiple addon at once
Copy-Item -Recurse testdata/servers/server_empty .debug-sandbox/testdata/servers/server_with_uninstalled_addons
Copy-Item -Recurse testdata/addons               .debug-sandbox/testdata/servers/server_with_uninstalled_addons/my-addons-collection

Write-Host "Sandbox ready!"
