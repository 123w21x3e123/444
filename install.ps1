$ErrorActionPreference = 'Stop'
$repo = '123w21x3e123/444'   # <- your GitHub user/repo
$dir  = Join-Path $env:LOCALAPPDATA '444'
New-Item -ItemType Directory -Force -Path $dir | Out-Null
Write-Host 'Downloading 444...'
Invoke-WebRequest -UseBasicParsing -Uri "https://github.com/$repo/releases/latest/download/444.exe" -OutFile (Join-Path $dir '444.exe')
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $dir) {
  [Environment]::SetEnvironmentVariable('Path', "$userPath;$dir", 'User')
  $env:Path += ";$dir"
}
Write-Host 'Installed. Close Spotify, then run:  444.exe apply'
