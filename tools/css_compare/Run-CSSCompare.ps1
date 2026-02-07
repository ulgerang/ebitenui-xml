<#
.SYNOPSIS
    CSS Visual Comparison Loop for EbitenUI-XML
    
.DESCRIPTION
    Automated loop that:
    1. Converts ebitenui-xml layout + styles → HTML reference page
    2. Launches ebitenui-xml app with ERTP debug server  
    3. Opens the HTML reference page in a headless browser
    4. Captures screenshots from both
    5. Performs pixel-level comparison
    6. Generates a visual diff report

.PARAMETER LayoutPath
    Path to the layout XML file

.PARAMETER StylesPath
    Path to the styles JSON file

.PARAMETER Width
    Canvas width (default: 640)

.PARAMETER Height
    Canvas height (default: 480)

.PARAMETER Port
    ERTP debug server port (default: 9222)

.PARAMETER OutputDir
    Directory for comparison output (default: ./css_compare_output)

.PARAMETER SkipBuild
    Skip building the Go binaries

.EXAMPLE
    .\Run-CSSCompare.ps1 -LayoutPath ../../assets/layout.xml -StylesPath ../../assets/styles.json
#>

param(
    [string]$LayoutPath = "../../assets/layout.xml",
    [string]$StylesPath = "../../assets/styles.json",
    [int]$Width = 640,
    [int]$Height = 480,
    [int]$Port = 9222,
    [string]$OutputDir = "./css_compare_output",
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Resolve-Path "$ScriptDir/../.."
$ERTPRoot = Resolve-Path "$ProjectRoot/../ebiten-ertp"

# ── Colors for output ──
function Write-Step { param($msg) Write-Host "▶ $msg" -ForegroundColor Cyan }
function Write-OK   { param($msg) Write-Host "✅ $msg" -ForegroundColor Green }
function Write-Warn { param($msg) Write-Host "⚠ $msg" -ForegroundColor Yellow }
function Write-Err  { param($msg) Write-Host "❌ $msg" -ForegroundColor Red }

# ── Setup ──
Write-Host ""
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Magenta
Write-Host "  CSS Visual Comparison Loop - EbitenUI-XML" -ForegroundColor Magenta  
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Magenta
Write-Host ""

# Create output directory
if (!(Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}
$OutputDir = Resolve-Path $OutputDir
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"

# Resolve input paths
$LayoutPath = Resolve-Path $LayoutPath
$StylesPath = Resolve-Path $StylesPath

Write-Step "Configuration:"
Write-Host "  Layout:    $LayoutPath"
Write-Host "  Styles:    $StylesPath"
Write-Host "  Size:      ${Width}x${Height}"
Write-Host "  ERTP Port: $Port"
Write-Host "  Output:    $OutputDir"
Write-Host ""

# ──────────────────────────────────────────────
# PHASE 1: Convert to HTML Reference
# ──────────────────────────────────────────────
Write-Step "Phase 1: Generating HTML reference page..."

$converterExe = "$ScriptDir/converter.exe"
$pixeldiffExe = "$ScriptDir/pixeldiff.exe"
if (!$SkipBuild -or !(Test-Path $converterExe) -or !(Test-Path $pixeldiffExe)) {
    Write-Host "  Building tools..."
    Push-Location $ProjectRoot
    go build -o "$ScriptDir/converter.exe" ./tools/css_compare/cmd/converter
    if ($LASTEXITCODE -ne 0) {
        Write-Err "Failed to build converter"
        Pop-Location
        exit 1
    }
    go build -o "$ScriptDir/pixeldiff.exe" ./tools/css_compare/cmd/pixeldiff
    if ($LASTEXITCODE -ne 0) {
        Write-Err "Failed to build pixeldiff"
        Pop-Location
        exit 1
    }
    Pop-Location
}

$referenceHtml = "$OutputDir/reference_${timestamp}.html"
& $converterExe -layout "$LayoutPath" -styles "$StylesPath" -out "$referenceHtml" -width $Width -height $Height
if ($LASTEXITCODE -ne 0) {
    Write-Err "Failed to generate HTML reference"
    exit 1
}
Write-OK "Reference HTML: $referenceHtml"

# ──────────────────────────────────────────────
# PHASE 2: Capture Browser Screenshot
# ──────────────────────────────────────────────
Write-Step "Phase 2: Capturing browser screenshot..."

$browserScreenshot = "$OutputDir/browser_${timestamp}.png"

# Try to use Playwright (if available) or fall back to Chrome/Edge headless
$chromePaths = @(
    "${env:ProgramFiles}\Google\Chrome\Application\chrome.exe",
    "${env:ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe",
    "${env:ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe",
    "${env:ProgramFiles}\Microsoft\Edge\Application\msedge.exe"
)

$chromeBin = $null
foreach ($path in $chromePaths) {
    if (Test-Path $path) {
        $chromeBin = $path
        break
    }
}

if ($chromeBin) {
    Write-Host "  Using browser: $chromeBin"
    
    # Convert file path to file:// URL 
    $fileUrl = "file:///" + ($referenceHtml -replace '\\','/')

    # Chrome headless screenshot
    $chromeArgs = @(
        "--headless=new",
        "--disable-gpu",
        "--no-sandbox",
        "--hide-scrollbars",
        "--window-size=${Width},${Height}",
        "--screenshot=$browserScreenshot",
        "$fileUrl"
    )
    
    & $chromeBin @chromeArgs 2>$null
    
    if (Test-Path $browserScreenshot) {
        Write-OK "Browser screenshot: $browserScreenshot"
    } else {
        Write-Warn "Browser screenshot failed, will skip pixel comparison"
    }
} else {
    Write-Warn "No Chrome/Edge found. Install Chrome for automated browser screenshots."
    Write-Host "  Manual step: Open $referenceHtml in browser and capture screenshot"
}

# ──────────────────────────────────────────────
# PHASE 3: Launch Ebiten App & Capture via ERTP
# ──────────────────────────────────────────────
Write-Step "Phase 3: Launching EbitenUI-XML app with ERTP..."

$ebitenScreenshot = "$OutputDir/ebiten_${timestamp}.png"

# Build the css_compare harness
if (!$SkipBuild) {
    Write-Host "  Building css_compare harness..."
    Push-Location $ProjectRoot
    go build -o "$OutputDir/css_compare.exe" ./cmd/css_compare
    if ($LASTEXITCODE -ne 0) {
        Write-Warn "Failed to build css_compare harness (need ebiten-ertp dependency)"
        Write-Host "  Will use existing demo if available..."
        Pop-Location
    } else {
        Pop-Location
    }
}

$harness = "$OutputDir/css_compare.exe"
if (Test-Path $harness) {
    # Launch harness in background
    $process = Start-Process -FilePath $harness `
        -ArgumentList "-layout", "$LayoutPath", "-styles", "$StylesPath", "-width", $Width, "-height", $Height, "-port", ":$Port" `
        -PassThru -WindowStyle Normal

    # Wait for ERTP to be ready
    Write-Host "  Waiting for ERTP server..."
    $ready = $false
    for ($i = 0; $i -lt 30; $i++) {
        Start-Sleep -Milliseconds 500
        try {
            $state = Invoke-RestMethod -Uri "http://localhost:$Port/state" -Method Get -TimeoutSec 2
            if ($state) {
                $ready = $true
                break
            }
        } catch {
            # Retry
        }
    }

    if ($ready) {
        Write-OK "ERTP connected: Tick=$($state.tick)"
        
        # Wait a few more frames for rendering to stabilize
        Start-Sleep -Seconds 1
        
        # Capture screenshot via ERTP
        Invoke-WebRequest -Uri "http://localhost:$Port/screenshot" -OutFile $ebitenScreenshot
        Write-OK "Ebiten screenshot: $ebitenScreenshot"
    } else {
        Write-Err "ERTP server did not respond within timeout"
    }

    # Stop the harness
    if (!$process.HasExited) {
        $process.Kill()
        $process.WaitForExit(3000)
    }
} else {
    Write-Warn "No harness executable found. Build manually:"
    Write-Host "  cd $ProjectRoot && go build -o ./tools/css_compare/css_compare_output/css_compare.exe ./cmd/css_compare"
}

# ──────────────────────────────────────────────
# PHASE 4: Generate Comparison Report
# ──────────────────────────────────────────────
Write-Step "Phase 4: Generating comparison report..."

$reportPath = "$OutputDir/report_${timestamp}.html"

$hasBrowser = Test-Path $browserScreenshot
$hasEbiten = Test-Path $ebitenScreenshot

# Generate pixel diff if both screenshots exist
$diffStats = $null
if ($hasBrowser -and $hasEbiten) {
    Write-Host "  Computing pixel differences..."
    $diffImage = "$OutputDir/diff_${timestamp}.png"
    
    $diffOutput = & $pixeldiffExe "$browserScreenshot" "$ebitenScreenshot" "$diffImage" 2>&1
    
    if (Test-Path $diffImage) {
        # Parse diff stats
        $diffStats = @{}
        foreach ($line in ($diffOutput -split "`n")) {
            if ($line -match "^(\w+)=(.+)$") {
                $diffStats[$matches[1]] = $matches[2].Trim()
            }
        }
        Write-OK "Diff image: $diffImage"
        Write-Host "  Diff Pixels: $($diffStats['DIFF_PIXELS']) / $($diffStats['TOTAL_PIXELS']) ($($diffStats['DIFF_PCT'])%)"
        Write-Host "  Avg Delta:   $($diffStats['AVG_DELTA'])"
    }
}

# ──────────────────────────────────────────────
# Generate HTML Report
# ──────────────────────────────────────────────

$browserImgTag = if ($hasBrowser) { "<img src='$(Split-Path $browserScreenshot -Leaf)' alt='Browser'>" } else { "<p class='no-data'>No browser screenshot</p>" }
$ebitenImgTag = if ($hasEbiten) { "<img src='$(Split-Path $ebitenScreenshot -Leaf)' alt='Ebiten'>" } else { "<p class='no-data'>No Ebiten screenshot</p>" }
$diffImgTag = if ($diffStats -and (Test-Path $diffImage)) { "<img src='$(Split-Path $diffImage -Leaf)' alt='Diff'>" } else { "<p class='no-data'>No diff available</p>" } 

$diffPct = if ($diffStats) { $diffStats['DIFF_PCT'] } else { "N/A" }
$diffCount = if ($diffStats) { $diffStats['DIFF_PIXELS'] } else { "N/A" }
$totalCount = if ($diffStats) { $diffStats['TOTAL_PIXELS'] } else { "N/A" }
$avgDelta = if ($diffStats) { $diffStats['AVG_DELTA'] } else { "N/A" }

$scoreColor = "green"
if ($diffStats -and [double]$diffStats['DIFF_PCT'] -gt 5) { $scoreColor = "orange" }
if ($diffStats -and [double]$diffStats['DIFF_PCT'] -gt 20) { $scoreColor = "red" }

$reportContent = @"
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>CSS Compare Report - $timestamp</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', system-ui, sans-serif; background: #0d1117; color: #c9d1d9; padding: 24px; }
        h1 { color: #58a6ff; margin-bottom: 8px; font-size: 24px; }
        .subtitle { color: #8b949e; margin-bottom: 24px; }
        
        .stats {
            display: flex; gap: 16px; margin-bottom: 24px;
        }
        .stat-card {
            background: #161b22; border: 1px solid #30363d; border-radius: 8px;
            padding: 16px 24px; flex: 1; text-align: center;
        }
        .stat-card .value { font-size: 28px; font-weight: bold; }
        .stat-card .label { font-size: 12px; color: #8b949e; margin-top: 4px; }
        
        .comparison {
            display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 16px;
            margin-bottom: 24px;
        }
        .panel {
            background: #161b22; border: 1px solid #30363d; border-radius: 8px;
            padding: 16px; overflow: hidden;
        }
        .panel h2 { font-size: 14px; color: #58a6ff; margin-bottom: 12px; text-transform: uppercase; letter-spacing: 1px; }
        .panel img { width: 100%; border-radius: 4px; border: 1px solid #30363d; }
        .panel .no-data { color: #8b949e; font-style: italic; padding: 40px 0; text-align: center; }
        
        .style-audit {
            background: #161b22; border: 1px solid #30363d; border-radius: 8px;
            padding: 24px; margin-bottom: 24px;
        }
        .style-audit h2 { color: #58a6ff; margin-bottom: 16px; }
        
        .properties-table { width: 100%; border-collapse: collapse; }
        .properties-table th { text-align: left; padding: 8px 12px; border-bottom: 1px solid #30363d; color: #8b949e; font-size: 12px; }
        .properties-table td { padding: 8px 12px; border-bottom: 1px solid #21262d; font-family: monospace; font-size: 13px; }
        .properties-table tr:hover { background: #1c2128; }
        
        .tag-impl { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 600; }
        .tag-yes { background: #238636; color: #fff; }
        .tag-partial { background: #9e6a03; color: #fff; }
        .tag-no { background: #da3633; color: #fff; }
        
        footer { text-align: center; color: #484f58; font-size: 12px; margin-top: 32px; }
    </style>
</head>
<body>
    <h1>🔍 CSS Visual Comparison Report</h1>
    <p class="subtitle">Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') | Layout: $(Split-Path $LayoutPath -Leaf) | Size: ${Width}x${Height}</p>
    
    <div class="stats">
        <div class="stat-card">
            <div class="value" style="color: $scoreColor">${diffPct}%</div>
            <div class="label">Pixel Difference</div>
        </div>
        <div class="stat-card">
            <div class="value">${diffCount}</div>
            <div class="label">Different Pixels</div>
        </div>
        <div class="stat-card">
            <div class="value">${totalCount}</div>
            <div class="label">Total Pixels</div>
        </div>
        <div class="stat-card">
            <div class="value">${avgDelta}</div>
            <div class="label">Avg Color Delta</div>
        </div>
    </div>
    
    <div class="comparison">
        <div class="panel">
            <h2>🌐 Browser (HTML/CSS Reference)</h2>
            $browserImgTag
        </div>
        <div class="panel">
            <h2>🎮 Ebiten (EbitenUI-XML)</h2>
            $ebitenImgTag
        </div>
        <div class="panel">
            <h2>🔴 Pixel Difference</h2>
            $diffImgTag
        </div>
    </div>
    
    <div class="style-audit">
        <h2>📋 CSS Property Implementation Audit</h2>
        <table class="properties-table">
            <thead>
                <tr>
                    <th>CSS Property</th>
                    <th>Status</th>
                    <th>Notes</th>
                </tr>
            </thead>
            <tbody>
                <tr><td>display: flex</td><td><span class="tag-impl tag-yes">YES</span></td><td>Core layout engine</td></tr>
                <tr><td>flex-direction</td><td><span class="tag-impl tag-yes">YES</span></td><td>row / column</td></tr>
                <tr><td>justify-content</td><td><span class="tag-impl tag-yes">YES</span></td><td>start, center, end, space-between, space-around, space-evenly</td></tr>
                <tr><td>align-items</td><td><span class="tag-impl tag-yes">YES</span></td><td>start, center, end, stretch</td></tr>
                <tr><td>flex-grow</td><td><span class="tag-impl tag-yes">YES</span></td><td>Distributes remaining space</td></tr>
                <tr><td>flex-wrap</td><td><span class="tag-impl tag-yes">YES</span></td><td>nowrap, wrap, wrap-reverse</td></tr>
                <tr><td>gap</td><td><span class="tag-impl tag-yes">YES</span></td><td>Spacing between flex children</td></tr>
                <tr><td>padding</td><td><span class="tag-impl tag-yes">YES</span></td><td>All four sides</td></tr>
                <tr><td>margin</td><td><span class="tag-impl tag-yes">YES</span></td><td>All four sides</td></tr>
                <tr><td>width / height</td><td><span class="tag-impl tag-yes">YES</span></td><td>Explicit sizing</td></tr>
                <tr><td>min/max-width/height</td><td><span class="tag-impl tag-yes">YES</span></td><td>Size constraints</td></tr>
                <tr><td>background (solid)</td><td><span class="tag-impl tag-yes">YES</span></td><td>Hex, RGB, RGBA, named colors</td></tr>
                <tr><td>background (gradient)</td><td><span class="tag-impl tag-yes">YES</span></td><td>linear-gradient, radial-gradient</td></tr>
                <tr><td>color</td><td><span class="tag-impl tag-yes">YES</span></td><td>Text color</td></tr>
                <tr><td>border</td><td><span class="tag-impl tag-yes">YES</span></td><td>Width + color</td></tr>
                <tr><td>border-radius</td><td><span class="tag-impl tag-yes">YES</span></td><td>Rounded corners</td></tr>
                <tr><td>box-shadow</td><td><span class="tag-impl tag-yes">YES</span></td><td>offset, blur, spread, color, inset</td></tr>
                <tr><td>text-shadow</td><td><span class="tag-impl tag-partial">PARTIAL</span></td><td>Basic support</td></tr>
                <tr><td>font-size</td><td><span class="tag-impl tag-yes">YES</span></td><td>Pixel-based sizing</td></tr>
                <tr><td>text-align</td><td><span class="tag-impl tag-yes">YES</span></td><td>left, center, right</td></tr>
                <tr><td>line-height</td><td><span class="tag-impl tag-yes">YES</span></td><td>Pixel units</td></tr>
                <tr><td>opacity</td><td><span class="tag-impl tag-yes">YES</span></td><td>0-1 float</td></tr>
                <tr><td>:hover</td><td><span class="tag-impl tag-yes">YES</span></td><td>Hover state styles</td></tr>
                <tr><td>:active</td><td><span class="tag-impl tag-yes">YES</span></td><td>Active/pressed state</td></tr>
                <tr><td>:disabled</td><td><span class="tag-impl tag-yes">YES</span></td><td>Disabled state</td></tr>
                <tr><td>:focus</td><td><span class="tag-impl tag-yes">YES</span></td><td>Focus state</td></tr>
                <tr><td>overflow (scroll)</td><td><span class="tag-impl tag-yes">YES</span></td><td>Scrollable container</td></tr>
                <tr><td>transform</td><td><span class="tag-impl tag-partial">PARTIAL</span></td><td>translate, scale, rotate, skew</td></tr>
                <tr><td>transition</td><td><span class="tag-impl tag-partial">PARTIAL</span></td><td>Property animations</td></tr>
                <tr><td>CSS Variables</td><td><span class="tag-impl tag-yes">YES</span></td><td>--var-name / var(--var-name)</td></tr>
                <tr><td>outline</td><td><span class="tag-impl tag-partial">PARTIAL</span></td><td>Basic outline support</td></tr>
                <tr><td>position absolute</td><td><span class="tag-impl tag-partial">PARTIAL</span></td><td>Limited positioning</td></tr>
                <tr><td>z-index</td><td><span class="tag-impl tag-yes">YES</span></td><td>Layering order</td></tr>
                <tr><td>font-family</td><td><span class="tag-impl tag-no">NO</span></td><td>Uses bitmap font only</td></tr>
                <tr><td>font-weight</td><td><span class="tag-impl tag-no">NO</span></td><td>Not supported (bitmap)</td></tr>
                <tr><td>text-decoration</td><td><span class="tag-impl tag-no">NO</span></td><td>Not implemented</td></tr>
                <tr><td>backdrop-filter</td><td><span class="tag-impl tag-no">NO</span></td><td>Blur/glassmorphism not rendered</td></tr>
                <tr><td>cursor</td><td><span class="tag-impl tag-no">NO</span></td><td>No cursor change in Ebiten</td></tr>
                <tr><td>overflow-x / overflow-y</td><td><span class="tag-impl tag-no">NO</span></td><td>Only combined overflow</td></tr>
            </tbody>
        </table>
    </div>
    
    <footer>
        Generated by CSS Compare Tool | EbitenUI-XML + ERTP
    </footer>
</body>
</html>
"@

Set-Content -Path $reportPath -Value $reportContent -Encoding UTF8
Write-OK "Report: $reportPath"

# ──────────────────────────────────────────────
# Summary
# ──────────────────────────────────────────────
Write-Host ""
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Magenta
Write-Host "  Comparison Complete!" -ForegroundColor Magenta
Write-Host "═══════════════════════════════════════════════" -ForegroundColor Magenta
Write-Host ""

if ($diffStats) {
    $pct = [double]$diffStats['DIFF_PCT']
    if ($pct -lt 1) {
        Write-OK "Excellent match! ($($diffStats['DIFF_PCT'])% difference)"
    } elseif ($pct -lt 5) {
        Write-Warn "Good match with minor differences ($($diffStats['DIFF_PCT'])%)"
    } elseif ($pct -lt 20) {
        Write-Warn "Moderate differences detected ($($diffStats['DIFF_PCT'])%)"
    } else {
        Write-Err "Significant differences detected ($($diffStats['DIFF_PCT'])%)"
    }
}

Write-Host ""
Write-Host "  📄 Report: $reportPath" -ForegroundColor White
Write-Host "  🌐 Reference: $referenceHtml" -ForegroundColor White
if ($hasBrowser) { Write-Host "  📸 Browser: $browserScreenshot" -ForegroundColor White }
if ($hasEbiten) { Write-Host "  🎮 Ebiten: $ebitenScreenshot" -ForegroundColor White }
Write-Host ""
Write-Host "  Open the report in browser: start $reportPath" -ForegroundColor Gray
