$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = Convert-Path("$PSScriptRoot/../..")

# Define the target binary file path based on the environment
$Target="$ProjectRoot\bin\bedrock_helper.exe"
if ($env:CI -eq "true") {
    $Target="$ProjectRoot\bin\bedrock_helper-$env:GOOS-$env:GOARCH.exe"
}

# Show compiled version
& $Target --version
