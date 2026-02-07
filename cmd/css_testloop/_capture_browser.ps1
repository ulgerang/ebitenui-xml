$ref = (Resolve-Path "cmd\css_testloop\output\reference.html").Path
$url = "file:///" + ($ref -replace '\\','/')
$out = Join-Path (Get-Location) "cmd\css_testloop\output\browser.png"

$chromePaths = @(
    "C:\Program Files\Google\Chrome\Application\chrome.exe",
    "C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe",
    "C:\Program Files\Microsoft\Edge\Application\msedge.exe"
)
$chrome = $null
foreach ($p in $chromePaths) { if (Test-Path $p) { $chrome = $p; break } }

if ($chrome) {
    Write-Host "Browser: $chrome"
    Write-Host "URL: $url"
    & $chrome --headless=new --disable-gpu --no-sandbox --hide-scrollbars --window-size=1070,850 "--screenshot=$out" $url 2>$null
    Start-Sleep -Seconds 2
    if (Test-Path $out) { Write-Host "OK: browser.png" } else { Write-Host "FAIL: no screenshot" }
} else {
    Write-Host "No browser found"
}
