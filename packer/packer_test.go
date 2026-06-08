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
		"[INFO] Packing: " + filepath.ToSlash(source),
		"[INFO] Packing: " + filepath.ToSlash(filepath.Join(source, "skins", "skin")),
		"[INFO] Packing: " + filepath.ToSlash(filepath.Join(source, "backgrounds", "background")),
		"[INFO] Packing: " + filepath.ToSlash(filepath.Join(source, "effects", "effect")),
		"[INFO] Packing: " + filepath.ToSlash(filepath.Join(source, "particles", "particle")),
		"[INFO] Packing: " + filepath.ToSlash(filepath.Join(source, "engines", "engine")),
		"[INFO] Packing: " + filepath.ToSlash(filepath.Join(source, "levels", "level")),
	} {
		if !bytes.Contains(log.Bytes(), []byte(want)) {
			t.Fatalf("log = %q; want %q", log.String(), want)
		}
	}
	if bytes.Contains(log.Bytes(), []byte("\\")) {
		t.Fatalf("log = %q; want slash-normalized paths", log.String())
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

func TestPackMissingItemDirectoriesAsEmptyArrays(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	output := filepath.Join(dir, "pack")
	mkdir(t, source)
	write(t, filepath.Join(source, "info.json"), `{"title":{"en":"Server"}}`)

	if err := packer.Pack(context.Background(), packer.Options{
		Input:  source,
		Output: output,
	}); err != nil {
		t.Fatal(err)
	}

	assertEmptySections(t, output)
}

func TestPackItemDirectoriesWithoutChildDirectoriesAsEmptyArrays(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	output := filepath.Join(dir, "pack")
	mkdir(t, source)
	write(t, filepath.Join(source, "info.json"), `{"title":{"en":"Server"}}`)
	for _, section := range itemSections() {
		mkdir(t, filepath.Join(source, section))
		write(t, filepath.Join(source, section, "ignored.txt"), "ignored")
	}

	if err := packer.Pack(context.Background(), packer.Options{
		Input:  source,
		Output: output,
	}); err != nil {
		t.Fatal(err)
	}

	assertEmptySections(t, output)
}

func TestPackPreservesPresentEmptyDescription(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	output := filepath.Join(dir, "pack")
	createMinimalSource(t, source)
	write(t, filepath.Join(source, "info.json"), `{"title":{"en":"Server"},"description":{}}`)
	writeItem(t, source, "posts", "post", `{"version":1,"title":{"en":"Post"},"time":1,"author":{"en":"Author"},"tags":[],"description":{}}`)

	if err := packer.Pack(context.Background(), packer.Options{
		Input:  source,
		Output: output,
	}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(output, "db.json"))
	if err != nil {
		t.Fatal(err)
	}
	var db map[string]json.RawMessage
	if err := json.Unmarshal(data, &db); err != nil {
		t.Fatal(err)
	}

	var info map[string]json.RawMessage
	if err := json.Unmarshal(db["info"], &info); err != nil {
		t.Fatal(err)
	}
	if got := string(info["description"]); got != "{}" {
		t.Fatalf("info.description = %s; want {} in %s", got, data)
	}

	var posts []map[string]json.RawMessage
	if err := json.Unmarshal(db["posts"], &posts); err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("posts length = %d; want 1 in %s", len(posts), data)
	}
	if got := string(posts[0]["description"]); got != "{}" {
		t.Fatalf("post.description = %s; want {} in %s", got, data)
	}

	var levels []map[string]json.RawMessage
	if err := json.Unmarshal(db["levels"], &levels); err != nil {
		t.Fatal(err)
	}
	if _, ok := levels[0]["description"]; ok {
		t.Fatalf("missing level description was output: %s", data)
	}
}

func TestPackPreservesPresentEmptyTagFields(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	output := filepath.Join(dir, "pack")
	createMinimalSource(t, source)
	writeItem(t, source, "posts", "post", `{"version":1,"title":{"en":"Post"},"time":1,"author":{"en":"Author"},"tags":[{"title":{},"icon":""},{}]}`)

	if err := packer.Pack(context.Background(), packer.Options{
		Input:  source,
		Output: output,
	}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(output, "db.json"))
	if err != nil {
		t.Fatal(err)
	}
	var db map[string]json.RawMessage
	if err := json.Unmarshal(data, &db); err != nil {
		t.Fatal(err)
	}
	var posts []map[string]json.RawMessage
	if err := json.Unmarshal(db["posts"], &posts); err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("posts length = %d; want 1 in %s", len(posts), data)
	}
	var tags []map[string]json.RawMessage
	if err := json.Unmarshal(posts[0]["tags"], &tags); err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 {
		t.Fatalf("tags length = %d; want 2 in %s", len(tags), data)
	}
	if got := string(tags[0]["title"]); got != "{}" {
		t.Fatalf("tag title = %s; want {} in %s", got, data)
	}
	if got := string(tags[0]["icon"]); got != `""` {
		t.Fatalf("tag icon = %s; want empty string in %s", got, data)
	}
	if _, ok := tags[1]["title"]; ok {
		t.Fatalf("missing tag title was output: %s", data)
	}
	if _, ok := tags[1]["icon"]; ok {
		t.Fatalf("missing tag icon was output: %s", data)
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

	err := packer.Pack(context.Background(), packer.Options{
		Input:  source,
		Output: output,
	})
	if err == nil {
		t.Fatal("expected reference error")
	}
	if got, want := err.Error(), "playlists/playlist: missing not found (/levels/0)"; got != want {
		t.Fatalf("error = %q; want %q", got, want)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("output should be removed, stat err: %v", err)
	}
}

func TestPackOutputDirectoryErrorUsesSlashNormalizedPath(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	parent := filepath.Join(dir, "parent")
	output := filepath.Join(parent, "pack")
	mkdir(t, source)
	write(t, filepath.Join(source, "info.json"), `{"title":{"en":"Server"}}`)
	write(t, parent, "not a directory")

	err := packer.Pack(context.Background(), packer.Options{
		Input:  source,
		Output: output,
	})
	if err == nil {
		t.Fatal("expected output directory error")
	}
	want := filepath.ToSlash(output)
	if !bytes.Contains([]byte(err.Error()), []byte(want)) {
		t.Fatalf("error = %q; want path %q", err.Error(), want)
	}
	if bytes.Contains([]byte(err.Error()), []byte("\\")) {
		t.Fatalf("error = %q; want slash-normalized path", err.Error())
	}
}

func assertEmptySections(t *testing.T, output string) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(output, "db.json"))
	if err != nil {
		t.Fatal(err)
	}
	var db map[string]json.RawMessage
	if err := json.Unmarshal(data, &db); err != nil {
		t.Fatal(err)
	}
	for _, section := range itemSections() {
		raw, ok := db[section]
		if !ok {
			t.Fatalf("%s missing from db.json: %s", section, data)
		}
		if string(raw) != "[]" {
			t.Fatalf("%s = %s; want [] in %s", section, raw, data)
		}
	}
}

func itemSections() []string {
	return []string{
		"posts",
		"playlists",
		"levels",
		"skins",
		"backgrounds",
		"effects",
		"particles",
		"engines",
		"replays",
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
