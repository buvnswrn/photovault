<#
.SYNOPSIS
  Create labels, milestones and issues from scripts/issues.json. Safe to re-run: existing items are skipped.
.EXAMPLE
  ./scripts/create-issues.ps1 -Repo buvnswrn/photovault -DryRun
#>
param(
    [string]$Repo = 'buvnswrn/photovault',
    [string]$Gh = 'gh',
    [switch]$DryRun
)
$ErrorActionPreference = 'Stop'
$data = Get-Content -Raw -Encoding UTF8 (Join-Path $PSScriptRoot 'issues.json') | ConvertFrom-Json
$adrBase = "https://github.com/$Repo/blob/main/docs/adr"
$adrFiles = @{}
Get-ChildItem (Join-Path $PSScriptRoot '..\docs\adr') -Filter '0*.md' | ForEach-Object { $adrFiles[$_.Name.Substring(0,4)] = $_.Name }

function Invoke-Gh { param([string[]]$GhArgs)
    if ($DryRun) { Write-Host "[dry-run] gh $($GhArgs -join ' ')"; return '' }
    $out = & $Gh @GhArgs
    if ($LASTEXITCODE -ne 0) { throw "gh $($GhArgs -join ' ') failed ($LASTEXITCODE)" }
    return $out
}

# Labels (--force updates existing ones)
foreach ($l in $data.labels) {
    Invoke-Gh @('label', 'create', $l.name, '--repo', $Repo, '--color', $l.color, '--description', "$($l.description) ", '--force') | Out-Null
}

# Milestones
$existingMs = @()
if (-not $DryRun) { $existingMs = (& $Gh api "repos/$Repo/milestones?state=all&per_page=100" | ConvertFrom-Json) | ForEach-Object { $_.title } }
foreach ($m in $data.milestones) {
    if ($existingMs -contains $m.title) { continue }
    Invoke-Gh @('api', "repos/$Repo/milestones", '-f', "title=$($m.title)", '-f', "description=$($m.description)") | Out-Null
}

# Existing issues by title (for idempotency) and ID -> number map
$numbers = @{}
$existing = @{}
if (-not $DryRun) {
    (& $Gh issue list --repo $Repo --state all --limit 500 --json number,title | ConvertFrom-Json) | ForEach-Object { $existing[$_.title] = $_.number }
}

foreach ($i in $data.issues) {
    $title = "[$($i.id)] $($i.title)"
    if ($existing.ContainsKey($title)) { $numbers[$i.id] = $existing[$title]; Write-Host "skip  $title (#$($existing[$title]))"; continue }

    $deps = if ($i.depends.Count) {
        ($i.depends | ForEach-Object { if ($numbers.ContainsKey($_)) { "- #$($numbers[$_]) ($_)" } else { "- $_" } }) -join "`n"
    } else { '- none' }
    $adrs = if ($i.adrs.Count) {
        ($i.adrs | ForEach-Object { if ($adrFiles.ContainsKey($_)) { "- [ADR-$_]($adrBase/$($adrFiles[$_]))" } else { "- ADR-$_" } }) -join "`n"
    } else { '- none' }
    $ac = ($i.ac | ForEach-Object { "- [ ] $_" }) -join "`n"

    $body = @"
**ID:** ``$($i.id)`` · **Milestone:** $($i.milestone)

## Description
$($i.description)

## Acceptance criteria
$ac

## Depends on
$deps

## Design references
$adrs
- [Roadmap](https://github.com/$Repo/blob/main/docs/ROADMAP.md)
"@
    $bodyFile = New-TemporaryFile
    [IO.File]::WriteAllText($bodyFile.FullName, $body, (New-Object Text.UTF8Encoding $false))
    $ghArgs = @('issue', 'create', '--repo', $Repo, '--title', $title, '--body-file', $bodyFile.FullName, '--milestone', $i.milestone)
    foreach ($lab in $i.labels) { $ghArgs += @('--label', $lab) }
    $url = Invoke-Gh $ghArgs
    Remove-Item $bodyFile -Force
    if ($url -match '/issues/(\d+)') { $numbers[$i.id] = [int]$Matches[1] }
    Write-Host "created $title $url"
}
