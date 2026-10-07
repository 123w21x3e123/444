# 444
Lean Spicetify-style patcher for the desktop Spotify (Windows, spotify.com version).

Setup: push to github.com/123w21x3e123/444, tag a release (v0.2.0) so Actions builds 444.exe, then:
    iwr -useb https://raw.githubusercontent.com/123w21x3e123/444/main/install.ps1 | iex

Use (close Spotify first):
    444.exe apply        444.exe restore        444.exe devtools
    444.exe config accent "#ff5500"

In Spotify: Alt+Click hides any element, Alt+Shift+Z undoes, Alt+L opens synced lyrics.
Your own CSS: %APPDATA%\444\user.css, then run 444.exe apply.
After a Spotify update, just run 444.exe apply again.
