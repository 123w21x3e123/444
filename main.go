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

const version = "0.1.0"
const repo = "123w21x3e123/444" // <- your GitHub user/repo

type Config struct {
	MusicDir string `json:"music_dir"`
	Accent   string `json:"accent"`
	Theme    string `json:"theme"`
	Port     int    `json:"port"`
}

func dataDir() string {
	d, _ := os.UserConfigDir()
	p := filepath.Join(d, "444")
	os.MkdirAll(p, 0o755)
	return p
}

func loadConfig() Config {
	c := Config{Accent: "#1db954", Theme: "dark", Port: 4444}
	if b, err := os.ReadFile(filepath.Join(dataDir(), "config.json")); err == nil {
		json.Unmarshal(b, &c)
	}
	return c
}

func saveConfig(c Config) error {
	b, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(filepath.Join(dataDir(), "config.json"), b, 0o644)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"help"}
	}
	var err error
	switch args[0] {
	case "scan":
		err = cmdScan(args[1:])
	case "play":
		err = cmdPlay()
	case "config":
		err = cmdConfig(args[1:])
	case "update":
		err = cmdUpdate()
	case "uninstall":
		err = cmdUninstall()
	case "version":
		fmt.Println(version)
	default:
		fmt.Println("444 - tiny local music player\n\n  444 scan <folder>\n  444 play\n  444 config [key value]   keys: music_dir accent theme port\n  444 update\n  444 uninstall\n  444 version")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func cmdConfig(a []string) error {
	c := loadConfig()
	if len(a) == 2 {
		switch a[0] {
		case "music_dir":
			c.MusicDir = a[1]
		case "accent":
			c.Accent = a[1]
		case "theme":
			if a[1] != "dark" && a[1] != "light" {
				return fmt.Errorf("theme must be dark or light")
			}
			c.Theme = a[1]
		case "port":
			n, err := strconv.Atoi(a[1])
			if err != nil {
				return err
			}
			c.Port = n
		default:
			return fmt.Errorf("unknown key %q", a[0])
		}
		if err := saveConfig(c); err != nil {
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
	dir := filepath.Dir(exe)
	script := fmt.Sprintf(`$p=([Environment]::GetEnvironmentVariable('Path','User') -split ';') | Where-Object { $_ -and $_ -ne '%s' }; [Environment]::SetEnvironmentVariable('Path',($p -join ';'),'User'); Start-Sleep 2; Remove-Item -Recurse -Force '%s'`, dir, dir)
	os.RemoveAll(dataDir())
	exec.Command("powershell", "-NoProfile", "-Command", script).Start()
	fmt.Println("444 removed (settings and library index too)")
	return nil
}
