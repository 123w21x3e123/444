package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//go:embed web/444.css
var baseCSS []byte

//go:embed web/444.js
var baseJS []byte

func appsDir() (string, error) {
	d := filepath.Join(os.Getenv("APPDATA"), "Spotify", "Apps")
	if _, err := os.Stat(filepath.Join(d, "xpui.spa")); err != nil {
		return "", fmt.Errorf("can't find %s\\xpui.spa (use the Spotify from spotify.com, not the Microsoft Store one)", d)
	}
	return d, nil
}

func hasMarker(path string) bool {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name == "444.js" {
			return true
		}
	}
	return false
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func inject(b []byte) []byte {
	tag := []byte(`<link rel="stylesheet" href="444.css"><script defer src="444.js"></script>`)
	i := bytes.LastIndex(b, []byte("</head>"))
	if i < 0 {
		return append(append([]byte{}, b...), tag...)
	}
	out := append([]byte{}, b[:i]...)
	out = append(out, tag...)
	return append(out, b[i:]...)
}

func patchZip(src, dst string, css, js []byte) error {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zr.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(out)
	found := false
	for _, f := range zr.File {
		switch f.Name {
		case "444.css", "444.js":
			continue
		case "index.html":
			rc, err := f.Open()
			if err != nil {
				return err
			}
			b, _ := io.ReadAll(rc)
			rc.Close()
			w, _ := zw.Create("index.html")
			w.Write(inject(b))
			found = true
		default:
			if err := zw.Copy(f); err != nil {
				return err
			}
		}
	}
	if !found {
		return fmt.Errorf("index.html not found inside xpui.spa; Spotify's layout changed")
	}
	w, _ := zw.Create("444.css")
	w.Write(css)
	w, _ = zw.Create("444.js")
	w.Write(js)
	if err := zw.Close(); err != nil {
		return err
	}
	return out.Close()
}

func cmdApply() error {
	d, err := appsDir()
	if err != nil {
		return err
	}
	spa, bak := filepath.Join(d, "xpui.spa"), filepath.Join(d, "xpui.spa.backup")
	if _, err := os.Stat(filepath.Join(d, "xpui")); err == nil {
		fmt.Println("warning: an extracted Apps\\xpui folder exists (Spicetify?). Spotify may load that instead.")
	}
	if !hasMarker(spa) { // stock or freshly updated Spotify: refresh the backup
		if err := copyFile(spa, bak); err != nil {
			return err
		}
	}
	css := []byte(fmt.Sprintf(":root{--c444-accent:%s}\n", loadConfig().Accent))
	css = append(css, baseCSS...)
	if u, err := os.ReadFile(filepath.Join(dataDir(), "user.css")); err == nil {
		css = append(css, u...)
	}
	tmp := spa + ".tmp"
	if err := patchZip(bak, tmp, css, baseJS); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Remove(spa); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("close Spotify first (%v)", err)
	}
	if err := os.Rename(tmp, spa); err != nil {
		copyFile(bak, spa)
		return err
	}
	fmt.Println("applied. Start Spotify. Alt+Click hides an element, Alt+L opens lyrics.")
	return nil
}

func cmdRestore() error {
	d, err := appsDir()
	if err != nil {
		return err
	}
	spa, bak := filepath.Join(d, "xpui.spa"), filepath.Join(d, "xpui.spa.backup")
	if !hasMarker(spa) {
		fmt.Println("Spotify is already stock")
		return nil
	}
	if _, err := os.Stat(bak); err != nil {
		return fmt.Errorf("no backup found")
	}
	if err := os.Remove(spa); err != nil {
		return fmt.Errorf("close Spotify first (%v)", err)
	}
	if err := copyFile(bak, spa); err != nil {
		return err
	}
	fmt.Println("restored stock Spotify")
	return nil
}

func cmdDevtools() error {
	p := filepath.Join(os.Getenv("APPDATA"), "Spotify", "prefs")
	b, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("can't read %s (open Spotify once, then close it)", p)
	}
	if !bytes.Contains(b, []byte("app.enable-developer-mode=true")) {
		b = append(b, []byte("\napp.enable-developer-mode=true\n")...)
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return err
		}
	}
	fmt.Println("done. With Spotify closed before running this, start it and press Ctrl+Shift+I")
	return nil
}
