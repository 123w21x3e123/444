package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

const version = "0.2.0"
const repo = "123w21x3e123/444"

type Config struct {
	Accent string `json:"accent"`
}

func dataDir() string {
	d, _ := os.UserConfigDir()
	p := filepath.Join(d, "444")
	os.MkdirAll(p, 0o755)
	return p
}

func loadConfig() Config {
	c := Config{Accent: "#1db954"}
	if b, err := os.ReadFile(filepath.Join(dataDir(), "config.json")); err == nil {
		json.Unmarshal(b, &c)
	}
	return c
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"help"}
	}
	var err error
	switch args[0] {
	case "apply":
		err = cmdApply()
	case "restore":
		err = cmdRestore()
	case "devtools":
		err = cmdDevtools()
	case "config":
		err = cmdConfig(args[1:])
	case "update":
		err = cmdUpdate()
	case "uninstall":
		err = cmdUninstall()
	case "version":
		fmt.Println(version)
	default:
		fmt.Println("444 - lean Spotify patcher\n\n  444 apply        patch Spotify (close Spotify first)\n  444 restore      back to stock Spotify\n  444 devtools     enable Ctrl+Shift+I in Spotify\n  444 config accent #ff5500\n  444 update | uninstall | version\n\nExtra CSS: put it in %APPDATA%\\444\\user.css then run 444 apply")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func cmdConfig(a []string) error {
	c := loadConfig()
	if len(a) == 2 && a[0] == "accent" {
		h := a[1]
		if _, err := strconv.ParseUint(h[1:], 16, 32); h[0] != '#' || (len(h) != 4 && len(h) != 7) || err != nil {
			return fmt.Errorf("accent must look like #1db954")
		}
		c.Accent = h
		b, _ := json.MarshalIndent(c, "", "  ")
		if err := os.WriteFile(filepath.Join(dataDir(), "config.json"), b, 0o644); err != nil {
			return err
		}
	}
	b, _ := json.MarshalIndent(c, "", "  ")
	fmt.Println(string(b))
	return nil
}

func cmdUpdate() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	resp, err := http.Get("https://github.com/" + repo + "/releases/latest/download/444.exe")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	f, err := os.Create(exe + ".new")
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	f.Close()
	os.Remove(exe + ".old")
	if err := os.Rename(exe, exe+".old"); err != nil {
		return err
	}
	if err := os.Rename(exe+".new", exe); err != nil {
		return err
	}
	fmt.Println("updated")
	return nil
}

func cmdUninstall() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := cmdRestore(); err != nil {
		fmt.Println("note:", err)
	}
	dir := filepath.Dir(exe)
	script := fmt.Sprintf(`$p=([Environment]::GetEnvironmentVariable('Path','User') -split ';') | Where-Object { $_ -and $_ -ne '%s' }; [Environment]::SetEnvironmentVariable('Path',($p -join ';'),'User'); Start-Sleep 2; Remove-Item -Recurse -Force '%s'`, dir, dir)
	os.RemoveAll(dataDir())
	exec.Command("powershell", "-NoProfile", "-Command", script).Start()
	fmt.Println("444 removed")
	return nil
}
