# 444
Tiny local music player. One UI, no plugins.

## Setup
1. Replace `OWNER/REPO` in `main.go` and `install.ps1` with your GitHub user/repo.
2. Push to GitHub, then tag a release: `git tag v0.1.0 && git push --tags` (Actions builds 444.exe).
3. Install on any PC:
   `iwr -useb https://raw.githubusercontent.com/OWNER/REPO/main/install.ps1 | iex`

## Use
    444.exe scan C:\Music
    444.exe play
    444.exe config accent "#ff5500"
    444.exe config theme light
    444.exe update
    444.exe uninstall

Note: PowerShell reads a bare `444` as a number, so type `444.exe`.
