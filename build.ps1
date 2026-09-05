# Builds dist\TokenFinder.exe (windowless GUI app) and writes SHA256 checksums.
# Requires: Go 1.22+ and rsrc (installed automatically if missing).
#
#   .\build.ps1                # version from internal/ui/version.go
#   .\build.ps1 -Version 1.2.3 # override version baked into the binary
param(
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$rsrc = Join-Path (go env GOPATH) "bin\rsrc.exe"
if (-not (Test-Path $rsrc)) {
    Write-Host "rsrc not found, installing..."
    go install github.com/akavel/rsrc@latest
}

# Embed the Common Controls v6 manifest (required by lxn/walk).
& $rsrc -manifest app.manifest -o rsrc.syso

$ldflags = "-H windowsgui -s -w"
if ($Version -ne "") {
    $ldflags += " -X tokenfinder/internal/ui.Version=$Version"
}

New-Item -ItemType Directory -Force dist | Out-Null
go build -trimpath -ldflags $ldflags -o dist\TokenFinder.exe .
if (-not $?) { exit 1 }

$hash = (Get-FileHash dist\TokenFinder.exe -Algorithm SHA256).Hash.ToLower()
"$hash  TokenFinder.exe" | Out-File -Encoding ascii dist\SHA256SUMS.txt

Write-Host "Built dist\TokenFinder.exe"
Write-Host "SHA256: $hash"
