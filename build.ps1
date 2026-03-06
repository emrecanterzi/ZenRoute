$BinaryName = "zenroute"
$BinDir = "bin"

if (!(Test-Path -Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir | Out-Null
}

$ErrorActionPreference = "Stop"

if (Test-Path -Path ".env") {
    Copy-Item -Path ".env" -Destination "$BinDir\.env" -Force
}
if (Test-Path -Path "bypass-domains.txt") {
    Copy-Item -Path "bypass-domains.txt" -Destination "$BinDir\bypass-domains.txt" -Force
}

Write-Host "Building ZenRoute for Windows..."

try {
    go build -o "$BinDir\$BinaryName-windows.exe" ./cmd/zenroute
    Write-Host "Build successful! Executable is located at: $BinDir\$BinaryName-windows.exe" -ForegroundColor Green
} catch {
    Write-Host "Build failed. Ensure 'go' is installed and in your system PATH." -ForegroundColor Red
    exit 1
}
