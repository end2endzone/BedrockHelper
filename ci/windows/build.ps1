$ErrorActionPreference = "Stop"

# Get the projet's root directory
$ProjectRoot = "$PSScriptRoot/../.."

Set-Location $ProjectRoot/src

$Version = (Get-Content $ProjectRoot\VERSION).Trim()
$Commit = (git rev-parse --short HEAD).Trim()
$Date = (Get-Date).ToString("yyyy-MM-dd")
$CpuArchitecture = switch ((Get-CimInstance Win32_Processor).Architecture) {
    9       { "x64" }
    12      { "arm64" }
    0       { "386" }
    default { "unknown" }
}
$OsType = "windows"

# Point this to the exact package path where your main function lives
$Pkg = "main"

Write-Host "Building bedrock_helper-$OsType-$CpuArchitecture.exe version $Version..."

# Run build from the root, pointing to the main package directory
go build -ldflags "-X '$Pkg.Version=$Version' -X '$Pkg.CommitHash=$Commit' -X '$Pkg.BuildDate=$Date'" -o $ProjectRoot/bin/bedrock_helper-$OsType-$CpuArchitecture.exe ./cmd/bedrock_helper

Write-Host "Build complete!"
