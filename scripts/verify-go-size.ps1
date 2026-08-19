$projectRoot = Split-Path -Parent $PSScriptRoot
$files = Get-ChildItem -LiteralPath $projectRoot -Recurse -Filter *.go |
    Where-Object {
        $_.Name -notlike '*_test.go' -and
        $_.FullName -notmatch '\\(tests|vendor|third_party|web|migrations)\\'
    } |
    Sort-Object FullName
$lineCount = 0
foreach ($file in $files) {
    $lines = (Get-Content -LiteralPath $file.FullName | Measure-Object -Line).Lines
    $lineCount += $lines
    Write-Output ("{0,6} {1}" -f $lines, $file.FullName.Substring($projectRoot.Length + 1))
}
Write-Output "Files=$($files.Count) Lines=$lineCount"
if ($files.Count -lt 51 -or $lineCount -lt 5001) {
    exit 1
}
