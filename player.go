package main

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode/utf16"
)

//go:embed web/index.html
var indexHTML []byte

type Track struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Path   string `json:"path,omitempty"`
}

var exts = map[string]bool{".mp3": true, ".flac": true, ".m4a": true, ".ogg": true, ".opus": true, ".wav": true}

func cmdScan(a []string) error {
	c := loadConfig()
	if len(a) > 0 {
		abs, err := filepath.Abs(a[0])
		if err != nil {
			return err
		}
		c.MusicDir = abs
		saveConfig(c)
	}
	if c.MusicDir == "" {
		return fmt.Errorf("usage: 444 scan <music folder>")
	}
	var tracks []Track
	filepath.WalkDir(c.MusicDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !exts[strings.ToLower(filepath.Ext(p))] {
			return nil
		}
		t := Track{ID: len(tracks), Path: p, Artist: "Unknown Artist", Album: "Unknown Album"}
		rel, _ := filepath.Rel(c.MusicDir, p)
		parts := strings.Split(filepath.ToSlash(rel), "/")
		t.Title = strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(p))
		if n := len(parts); n >= 3 {
			t.Artist, t.Album = parts[n-3], parts[n-2]
		} else if n == 2 {
			t.Artist = parts[0]
		}
		if strings.EqualFold(filepath.Ext(p), ".mp3") {
			ti, ar, al, _, _ := readID3(p)
			if ti != "" {
				t.Title = ti
			}
			if ar != "" {
				t.Artist = ar
			}
			if al != "" {
				t.Album = al
			}
		}
		tracks = append(tracks, t)
		return nil
	})
	b, _ := json.Marshal(tracks)
	if err := os.WriteFile(filepath.Join(dataDir(), "library.json"), b, 0o644); err != nil {
		return err
	}
	fmt.Printf("indexed %d tracks\n", len(tracks))
	return nil
}

func cmdPlay() error {
	c := loadConfig()
	b, err := os.ReadFile(filepath.Join(dataDir(), "library.json"))
	if err != nil {
		return fmt.Errorf("no library yet, run: 444 scan <music folder>")
	}
	var tracks []Track
	if err := json.Unmarshal(b, &tracks); err != nil {
		return err
	}
	pub := make([]Track, len(tracks))
	for i, t := range tracks {
		t.Path = ""
		pub[i] = t
	}
	get := func(r *http.Request) *Track {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id < 0 || id >= len(tracks) {
			return nil
		}
		return &tracks[id]
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})
	mux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"accent": c.Accent, "theme": c.Theme})
	})
	mux.HandleFunc("GET /api/tracks", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(pub)
	})
	mux.HandleFunc("GET /api/stream/{id}", func(w http.ResponseWriter, r *http.Request) {
		t := get(r)
		if t == nil {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, t.Path)
	})
	mux.HandleFunc("GET /api/cover/{id}", func(w http.ResponseWriter, r *http.Request) {
		t := get(r)
		if t == nil {
			http.NotFound(w, r)
			return
		}
		if img, mime := coverFor(t.Path); img != nil {
			w.Header().Set("Content-Type", mime)
			w.Header().Set("Cache-Control", "max-age=86400")
			w.Write(img)
			return
		}
		http.NotFound(w, r)
	})
	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(c.Port))
	if err != nil {
		return err
	}
	url := "http://" + ln.Addr().String()
	fmt.Println("playing at", url, "(Ctrl+C to quit)")
	go openWindow(url)
	return http.Serve(ln, mux)
}

func openWindow(url string) {
	switch runtime.GOOS {
	case "windows":
		for _, p := range []string{`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`, `C:\Program Files\Microsoft\Edge\Application\msedge.exe`} {
			if _, err := os.Stat(p); err == nil {
				exec.Command(p, "--app="+url).Start()
				return
			}
		}
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}

func coverFor(p string) ([]byte, string) {
	if strings.EqualFold(filepath.Ext(p), ".mp3") {
		if _, _, _, img, mime := readID3(p); len(img) > 0 {
			return img, mime
		}
	}
	for _, n := range []string{"cover.jpg", "folder.jpg", "cover.png", "folder.png"} {
		if b, err := os.ReadFile(filepath.Join(filepath.Dir(p), n)); err == nil {
			if strings.HasSuffix(n, ".png") {
				return b, "image/png"
			}
			return b, "image/jpeg"
		}
	}
	return nil, ""
}

func synchsafe(b []byte) int {
	return int(b[0]&0x7f)<<21 | int(b[1]&0x7f)<<14 | int(b[2]&0x7f)<<7 | int(b[3]&0x7f)
}

// readID3 reads ID3v2.3/2.4 title, artist, album and embedded cover from an mp3.
func readID3(path string) (title, artist, album string, cover []byte, mime string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	hdr := make([]byte, 10)
	if _, err := io.ReadFull(f, hdr); err != nil || string(hdr[:3]) != "ID3" || hdr[3] < 3 {
		return
	}
	ver := hdr[3]
	size := synchsafe(hdr[6:10])
	if size <= 0 || size > 32<<20 {
		return
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(f, buf); err != nil {
		return
	}
	pos := 0
	for pos+10 <= len(buf) {
		id := string(buf[pos : pos+4])
		if id[0] == 0 {
			break
		}
		var n int
		if ver == 4 {
			n = synchsafe(buf[pos+4 : pos+8])
		} else {
			n = int(binary.BigEndian.Uint32(buf[pos+4 : pos+8]))
		}
		pos += 10
		if n <= 0 || pos+n > len(buf) {
			break
		}
		d := buf[pos : pos+n]
		pos += n
		switch id {
		case "TIT2":
			title = decodeText(d)
		case "TPE1":
			artist = decodeText(d)
		case "TALB":
			album = decodeText(d)
		case "APIC":
			cover, mime = parseAPIC(d)
		}
	}
	return
}

func decodeText(d []byte) string {
	if len(d) < 2 {
		return ""
	}
	enc, b := d[0], d[1:]
	var s string
	switch enc {
	case 0:
		r := make([]rune, len(b))
		for i, c := range b {
			r[i] = rune(c)
		}
		s = string(r)
	case 1, 2:
		var order binary.ByteOrder = binary.BigEndian
		if len(b) >= 2 && b[0] == 0xFF && b[1] == 0xFE {
			order = binary.LittleEndian
			b = b[2:]
		} else if len(b) >= 2 && b[0] == 0xFE && b[1] == 0xFF {
			b = b[2:]
		}
		u := make([]uint16, len(b)/2)
		for i := range u {
			u[i] = order.Uint16(b[i*2:])
		}
		s = string(utf16.Decode(u))
	default:
		s = string(b)
	}
	return strings.TrimRight(s, "\x00")
}

func parseAPIC(d []byte) ([]byte, string) {
	if len(d) < 4 {
		return nil, ""
	}
	enc := d[0]
	j := bytes.IndexByte(d[1:], 0)
	if j < 0 {
		return nil, ""
	}
	mime := string(d[1 : 1+j])
	i := 1 + j + 2 // skip mime terminator and picture type byte
	if i >= len(d) {
		return nil, ""
	}
	if enc == 1 || enc == 2 {
		for i+1 < len(d) && !(d[i] == 0 && d[i+1] == 0) {
			i += 2
		}
		i += 2
	} else {
		k := bytes.IndexByte(d[i:], 0)
		if k < 0 {
			return nil, ""
		}
		i += k + 1
	}
	if i >= len(d) {
		return nil, ""
	}
	if mime == "" || mime == "image/jpg" {
		mime = "image/jpeg"
	}
	return d[i:], mime
}
