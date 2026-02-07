<#
.SYNOPSIS
    CSS TestLoop - Automated visual comparison pipeline
.DESCRIPTION
    1. Build css_testloop
    2. Generate HTML reference
    3. Chrome headless screenshot
    4. Ebiten render + screenshot
    5. Compare + report
.EXAMPLE
    .\Run-CSSTestLoop.ps1
#>
param(
    [string]$OutputDir = "./output",
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path "$ScriptDir/../.."

function Write-Step { param($msg) Write-Host ">> $msg" -ForegroundColor Cyan }
function Write-OK   { param($msg) Write-Host "[OK] $msg" -ForegroundColor Green }
function Write-Err  { param($msg) Write-Host "[ERR] $msg" -ForegroundColor Red }

Write-Host ""
Write-Host "=== CSS TestLoop ===" -ForegroundColor Magenta
Write-Host ""

# Setup output dir
if (!(Test-Path $OutputDir)) { New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null }
$OutputDir = Resolve-Path $OutputDir

$exe = "$OutputDir/css_testloop.exe"

# Phase 1: Build
if (!$SkipBuild -or !(Test-Path $exe)) {
    Write-Step "Building css_testloop..."
    Push-Location $ProjectRoot
    go build -o $exe ./cmd/css_testloop
    if ($LASTEXITCODE -ne 0) { Write-Err "Build failed"; Pop-Location; exit 1 }
    Pop-Location
    Write-OK "Built: $exe"
}

# Phase 2: Generate HTML reference
Write-Step "Generating HTML reference..."
& $exe -mode html -out "$OutputDir/reference.html"
if ($LASTEXITCODE -ne 0) { Write-Err "HTML generation failed"; exit 1 }
Write-OK "reference.html"

# Phase 3: Chrome headless screenshot
Write-Step "Chrome headless screenshot..."
$chromePaths = @(
    "${env:ProgramFiles}\Google\Chrome\Application\chrome.exe",
    "${env:ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe",
    "${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe",
    "${env:ProgramFiles}\Microsoft\Edge\Application\msedge.exe"
)
$chromeBin = $null
foreach ($p in $chromePaths) { if (Test-Path $p) { $chromeBin = $p; break } }

$browserPng = "$OutputDir/browser.png"
if ($chromeBin) {
    $fileUrl = "file:///" + ($OutputDir -replace '\\','/') + "/reference.html"
    # Get page dimensions from reference to set viewport
    & $chromeBin --headless=new --disable-gpu --no-sandbox --hide-scrollbars --window-size=1200,1000 --screenshot="$browserPng" "$fileUrl" 2>$null
    if (Test-Path $browserPng) {
        Write-OK "browser.png"
    } else {
        Write-Err "Browser screenshot failed"
    }
} else {
    Write-Err "No Chrome/Edge found"
}

# Phase 4: Ebiten render + screenshot
Write-Step "Ebiten render..."
& $exe -mode render -out "$OutputDir/ebiten.png"
if (Test-Path "$OutputDir/ebiten.png") {
    Write-OK "ebiten.png"
} else {
    Write-Err "Ebiten screenshot failed"
}

# Phase 5: Compare
Write-Step "Comparing..."
if ((Test-Path $browserPng) -and (Test-Path "$OutputDir/ebiten.png")) {
    & $exe -mode compare -browser $browserPng -ebiten "$OutputDir/ebiten.png" -out "$OutputDir/report.html"
    Write-OK "report.html"
    Write-Host ""
    Write-Host "Open report: start $OutputDir\report.html" -ForegroundColor White
} else {
    Write-Err "Missing screenshots for comparison"
}

Write-Host ""
Write-Host "=== Done ===" -ForegroundColor Magenta
