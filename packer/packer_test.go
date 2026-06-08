package packer_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/WindowsSov8forUs/sonolus-pack-go/packer"
)

func TestPackMinimalSourceTree(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	output := filepath.Join(dir, "pack")
	createMinimalSource(t, source)
	var log bytes.Buffer

	if err := packer.Pack(context.Background(), packer.Options{
		Input:  source,
		Output: output,
		Logger: testLogger{w: &log},
	}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"[INFO] Packing: " + source,
		"[INFO] Packing: " + filepath.Join(source, "skins", "skin"),
		"[INFO] Packing: " + filepath.Join(source, "backgrounds", "background"),
		"[INFO] Packing: " + filepath.Join(source, "effects", "effect"),
		"[INFO] Packing: " + filepath.Join(source, "particles", "particle"),
		"[INFO] Packing: " + filepath.Join(source, "engines", "engine"),
		"[INFO] Packing: " + filepath.Join(source, "levels", "level"),
	} {
		if !bytes.Contains(log.Bytes(), []byte(want)) {
			t.Fatalf("log = %q; want %q", log.String(), want)
		}
	}

	if _, err := os.Stat(filepath.Join(output, "db.json")); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(output, "db.json"))
	if err != nil {
		t.Fatal(err)
	}
	var db map[string]any
	if err := json.Unmarshal(data, &db); err != nil {
		t.Fatal(err)
	}
	levels := db["levels"].([]any)
	level := levels[0].(map[string]any)
	if _, ok := level["meta"]; !ok {
		t.Fatalf("level meta was not preserved: %s", data)
	}
	entries, err := os.ReadDir(filepath.Join(output, "repository"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected repository files")
	}
}

func TestPackRemovesOutputOnReferenceError(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	output := filepath.Join(dir, "pack")
	mkdir(t, source)
	write(t, filepath.Join(source, "info.json"), `{"title":{"en":"Server"}}`)
	mkdir(t, filepath.Join(source, "playlists", "playlist"))
	write(t, filepath.Join(source, "playlists", "playlist", "item.json"), `{
		"version": 1,
		"title": {"en": "Playlist"},
		"subtitle": {"en": "Subtitle"},
		"author": {"en": "Author"},
		"tags": [],
		"levels": ["missing"]
	}`)

	if err := packer.Pack(context.Background(), packer.Options{
		Input:  source,
		Output: output,
	}); err == nil {
		t.Fatal("expected reference error")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("output should be removed, stat err: %v", err)
	}
}

func createMinimalSource(t *testing.T, source string) {
	t.Helper()

	mkdir(t, source)
	write(t, filepath.Join(source, "info.json"), `{"title":{"en":"Server"},"extra":"drop"}`)

	writeItem(t, source, "skins", "skin", `{"version":4,"title":{"en":"Skin"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[]}`)
	writeResourceSet(t, filepath.Join(source, "skins", "skin"), "thumbnail", "data", "texture")

	writeItem(t, source, "backgrounds", "background", `{"version":2,"title":{"en":"Background"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[]}`)
	writeResourceSet(t, filepath.Join(source, "backgrounds", "background"), "thumbnail", "data", "image", "configuration")

	writeItem(t, source, "effects", "effect", `{"version":5,"title":{"en":"Effect"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[]}`)
	writeResourceSet(t, filepath.Join(source, "effects", "effect"), "thumbnail", "data", "audio")

	writeItem(t, source, "particles", "particle", `{"version":3,"title":{"en":"Particle"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[]}`)
	writeResourceSet(t, filepath.Join(source, "particles", "particle"), "thumbnail", "data", "texture")

	writeItem(t, source, "engines", "engine", `{"version":13,"title":{"en":"Engine"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[],"skin":"skin","background":"background","effect":"effect","particle":"particle"}`)
	writeResourceSet(t, filepath.Join(source, "engines", "engine"), "thumbnail", "playData", "watchData", "previewData", "tutorialData", "configuration")

	writeItem(t, source, "levels", "level", `{"version":1,"rating":1,"title":{"en":"Level"},"artists":{"en":"Artist"},"author":{"en":"Author"},"tags":[],"engine":"engine","useSkin":{"useDefault":true},"useBackground":{"useDefault":true},"useEffect":{"useDefault":true},"useParticle":{"useDefault":true},"meta":{"keep":true}}`)
	writeResourceSet(t, filepath.Join(source, "levels", "level"), "cover", "bgm", "data")
}

func writeItem(t *testing.T, source, section, name, item string) {
	t.Helper()
	dir := filepath.Join(source, section, name)
	mkdir(t, dir)
	write(t, filepath.Join(dir, "item.json"), item)
}

func writeResourceSet(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, name := range names {
		write(t, filepath.Join(dir, name), name)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

type testLogger struct {
	w *bytes.Buffer
}

func (l testLogger) Info(args ...any) {
	fmt.Fprintln(l.w, append([]any{"[INFO]"}, args...)...)
}

func (l testLogger) Warning(args ...any) {
	fmt.Fprintln(l.w, append([]any{"[WARNING]"}, args...)...)
}
