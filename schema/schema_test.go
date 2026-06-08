package schema_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WindowsSov8forUs/sonolus-pack-go/schema"
)

func TestParsePostItemPreservesMetaAndCleansUnknownFields(t *testing.T) {
	path := writeTemp(t, `{
		"version": 1,
		"title": { "en": "Title" },
		"time": 1,
		"author": { "en": "Author" },
		"tags": [],
		"meta": { "keep": true },
		"extra": "drop"
	}`)

	item, err := schema.ParsePostItem(path)
	if err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["meta"]; !ok {
		t.Fatalf("meta was not preserved: %s", data)
	}
	if _, ok := fields["extra"]; ok {
		t.Fatalf("extra field was not cleaned: %s", data)
	}
}

func TestParseIgnoresUnknownFieldsWithKnownNamesFromOtherSchemas(t *testing.T) {
	infoPath := writeTemp(t, `{
		"title": { "en": "Server" },
		"hash": 1
	}`)
	info, err := schema.ParseInfo(infoPath)
	if err != nil {
		t.Fatal(err)
	}
	infoData, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(infoData), "hash") {
		t.Fatalf("info unknown hash was not cleaned: %s", infoData)
	}

	postPath := writeTemp(t, `{
		"version": 1,
		"title": { "en": "Title" },
		"time": 1,
		"author": { "en": "Author" },
		"tags": [],
		"levels": [1]
	}`)
	post, err := schema.ParsePostItem(postPath)
	if err != nil {
		t.Fatal(err)
	}
	postData, err := json.Marshal(post)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(postData), "levels") {
		t.Fatalf("post unknown levels was not cleaned: %s", postData)
	}
}

func TestParsePostItemRejectsWrongVersion(t *testing.T) {
	path := writeTemp(t, `{
		"version": 2,
		"title": { "en": "Title" },
		"time": 1,
		"author": { "en": "Author" },
		"tags": []
	}`)

	if _, err := schema.ParsePostItem(path); err == nil {
		t.Fatal("expected version error")
	}
}

func TestParsePostItemMissingVersionReportsOnlyRequired(t *testing.T) {
	path := writeTemp(t, `{
		"title": { "en": "Title" },
		"time": 1,
		"author": { "en": "Author" },
		"tags": []
	}`)

	_, err := schema.ParsePostItem(path)
	if err == nil {
		t.Fatal("expected version error")
	}
	var validationErrs *schema.ValidationErrors
	if !errors.As(err, &validationErrs) {
		t.Fatalf("error = %T %v; want ValidationErrors", err, err)
	}
	count := 0
	for _, item := range validationErrs.Items {
		if item.JSONPath == "/version" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("version error count = %d; want 1", count)
	}
}

func TestParsePostItemAcceptsFloatVersionLiteral(t *testing.T) {
	path := writeTemp(t, `{
		"version": 1.0,
		"title": { "en": "Title" },
		"time": 1,
		"author": { "en": "Author" },
		"tags": []
	}`)

	item, err := schema.ParsePostItem(path)
	if err != nil {
		t.Fatal(err)
	}
	if item.Version != 1 {
		t.Fatalf("version = %d; want 1", item.Version)
	}
}

func TestParsePostItemRejectsInvalidVersionTypes(t *testing.T) {
	tests := []string{
		`null`,
		`"1"`,
		`true`,
		`2`,
	}

	for _, version := range tests {
		t.Run(version, func(t *testing.T) {
			path := writeTemp(t, `{
				"version": `+version+`,
				"title": { "en": "Title" },
				"time": 1,
				"author": { "en": "Author" },
				"tags": []
			}`)
			if _, err := schema.ParsePostItem(path); err == nil {
				t.Fatal("expected version error")
			}
		})
	}
}

func TestParseLevelItemRejectsUseDefaultFalseWithoutItem(t *testing.T) {
	path := writeTemp(t, `{
		"version": 1,
		"rating": 1,
		"title": { "en": "Title" },
		"artists": { "en": "Artist" },
		"author": { "en": "Author" },
		"tags": [],
		"engine": "engine",
		"useSkin": { "useDefault": false },
		"useBackground": { "useDefault": true },
		"useEffect": { "useDefault": true },
		"useParticle": { "useDefault": true }
	}`)

	if _, err := schema.ParseLevelItem(path); err == nil {
		t.Fatal("expected useDefault false item error")
	}
}

func TestParseLevelItemCleansUseDefaultTrueItem(t *testing.T) {
	path := writeTemp(t, `{
		"version": 1,
		"rating": 1,
		"title": { "en": "Title" },
		"artists": { "en": "Artist" },
		"author": { "en": "Author" },
		"tags": [],
		"engine": "engine",
		"useSkin": { "useDefault": true, "item": 1 },
		"useBackground": { "useDefault": true },
		"useEffect": { "useDefault": true },
		"useParticle": { "useDefault": true }
	}`)

	item, err := schema.ParseLevelItem(path)
	if err != nil {
		t.Fatal(err)
	}
	if item.UseSkin.Item != "" {
		t.Fatalf("useDefault true item = %q; want cleaned empty", item.UseSkin.Item)
	}
	data, err := json.Marshal(item.UseSkin)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"useDefault":true}` {
		t.Fatalf("marshaled use item = %s; want cleaned item omitted", data)
	}
}

func TestParseLevelItemPreservesUseDefaultFalseEmptyItem(t *testing.T) {
	path := writeTemp(t, `{
		"version": 1,
		"rating": 1,
		"title": { "en": "Title" },
		"artists": { "en": "Artist" },
		"author": { "en": "Author" },
		"tags": [],
		"engine": "engine",
		"useSkin": { "useDefault": false, "item": "" },
		"useBackground": { "useDefault": true },
		"useEffect": { "useDefault": true },
		"useParticle": { "useDefault": true }
	}`)

	item, err := schema.ParseLevelItem(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(item.UseSkin)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"useDefault":false,"item":""}` {
		t.Fatalf("marshaled use item = %s; want item preserved", data)
	}
}

func TestParseCleansNestedUnknownFields(t *testing.T) {
	levelPath := writeTemp(t, `{
		"version": 1,
		"rating": 1,
		"title": { "en": "Title" },
		"artists": { "en": "Artist" },
		"author": { "en": "Author" },
		"tags": [{ "title": { "en": "Tag" }, "icon": "tag", "extra": "drop" }],
		"engine": "engine",
		"useSkin": { "useDefault": true, "item": "skin", "extra": "drop" },
		"useBackground": { "useDefault": true },
		"useEffect": { "useDefault": true },
		"useParticle": { "useDefault": true }
	}`)
	level, err := schema.ParseLevelItem(levelPath)
	if err != nil {
		t.Fatal(err)
	}
	levelData, err := json.Marshal(level)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(levelData), "extra") {
		t.Fatalf("nested extra field was not cleaned: %s", levelData)
	}
	if strings.Contains(string(levelData), `"item":"skin"`) {
		t.Fatalf("useDefault true item was not cleaned: %s", levelData)
	}

	srlPath := writeTemp(t, `{"hash":"hash","url":"url","extra":"drop"}`)
	srl, err := schema.ParseSrl(srlPath)
	if err != nil {
		t.Fatal(err)
	}
	srlData, err := json.Marshal(srl)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(srlData), "extra") {
		t.Fatalf("srl extra field was not cleaned: %s", srlData)
	}
}

func TestParseRejectsInvalidNumberFields(t *testing.T) {
	tests := []struct {
		name    string
		content string
		parse   func(string) error
	}{
		{
			name:    "post time null",
			content: `{"version":1,"title":{"en":"Title"},"time":null,"author":{"en":"Author"},"tags":[]}`,
			parse: func(path string) error {
				_, err := schema.ParsePostItem(path)
				return err
			},
		},
		{
			name:    "post time string",
			content: `{"version":1,"title":{"en":"Title"},"time":"1","author":{"en":"Author"},"tags":[]}`,
			parse: func(path string) error {
				_, err := schema.ParsePostItem(path)
				return err
			},
		},
		{
			name:    "level rating null",
			content: `{"version":1,"rating":null,"title":{"en":"Title"},"artists":{"en":"Artist"},"author":{"en":"Author"},"tags":[],"engine":"engine","useSkin":{"useDefault":true},"useBackground":{"useDefault":true},"useEffect":{"useDefault":true},"useParticle":{"useDefault":true}}`,
			parse: func(path string) error {
				_, err := schema.ParseLevelItem(path)
				return err
			},
		},
		{
			name:    "level rating object",
			content: `{"version":1,"rating":{},"title":{"en":"Title"},"artists":{"en":"Artist"},"author":{"en":"Author"},"tags":[],"engine":"engine","useSkin":{"useDefault":true},"useBackground":{"useDefault":true},"useEffect":{"useDefault":true},"useParticle":{"useDefault":true}}`,
			parse: func(path string) error {
				_, err := schema.ParseLevelItem(path)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.parse(writeTemp(t, tt.content)); err == nil {
				t.Fatal("expected number error")
			}
		})
	}
}

func TestParseReturnsMultipleValidationErrors(t *testing.T) {
	path := writeTemp(t, `{
		"version": "1",
		"title": { "en": 1 },
		"time": "1",
		"author": { "en": "Author" },
		"tags": []
	}`)

	_, err := schema.ParsePostItem(path)
	if err == nil {
		t.Fatal("expected validation errors")
	}
	var validationErrs *schema.ValidationErrors
	if !errors.As(err, &validationErrs) {
		t.Fatalf("error = %T %v; want ValidationErrors", err, err)
	}
	if len(validationErrs.Items) < 3 {
		t.Fatalf("validation error count = %d; want at least 3", len(validationErrs.Items))
	}
}

func TestParseMissingFileReportsDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "item.json")

	_, err := schema.ParsePostItem(path)
	if err == nil {
		t.Fatal("expected missing file error")
	}
	if !strings.Contains(err.Error(), "Does not exist") {
		t.Fatalf("error = %q; want Does not exist", err.Error())
	}
}

func TestParseReadErrorUsesSlashNormalizedPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "item.json")
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := schema.ParsePostItem(path)
	if err == nil {
		t.Fatal("expected read error")
	}
	want := filepath.ToSlash(path)
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error = %q; want path %q", err.Error(), want)
	}
	if strings.Contains(err.Error(), `\`) {
		t.Fatalf("error = %q; want slash-normalized path", err.Error())
	}
}

func TestParseRejectsTopLevelNonObjectsAsValidationErrors(t *testing.T) {
	tests := []string{
		`[]`,
		`"bad"`,
		`1`,
		`true`,
		`null`,
	}

	for _, content := range tests {
		t.Run(content, func(t *testing.T) {
			_, err := schema.ParsePostItem(writeTemp(t, content))
			if err == nil {
				t.Fatal("expected validation error")
			}
			var validationErrs *schema.ValidationErrors
			if !errors.As(err, &validationErrs) {
				t.Fatalf("error = %T %v; want ValidationErrors", err, err)
			}
		})
	}
}

func TestParseSrlAcceptsNullableFields(t *testing.T) {
	path := writeTemp(t, `{"hash": null, "url": "https://example.com"}`)

	if _, err := schema.ParseSrl(path); err != nil {
		t.Fatal(err)
	}
}

func TestParseRejectsInvalidFieldTypes(t *testing.T) {
	tests := []struct {
		name    string
		content string
		parse   func(string) error
	}{
		{
			name:    "localization value number",
			content: `{"version":1,"title":{"en":1},"time":1,"author":{"en":"Author"},"tags":[]}`,
			parse: func(path string) error {
				_, err := schema.ParsePostItem(path)
				return err
			},
		},
		{
			name:    "localization value null",
			content: `{"version":1,"title":{"en":null},"time":1,"author":{"en":"Author"},"tags":[]}`,
			parse: func(path string) error {
				_, err := schema.ParsePostItem(path)
				return err
			},
		},
		{
			name:    "tag non object",
			content: `{"version":1,"title":{"en":"Title"},"time":1,"author":{"en":"Author"},"tags":["bad"]}`,
			parse: func(path string) error {
				_, err := schema.ParsePostItem(path)
				return err
			},
		},
		{
			name:    "tag icon non string",
			content: `{"version":1,"title":{"en":"Title"},"time":1,"author":{"en":"Author"},"tags":[{"icon":1}]}`,
			parse: func(path string) error {
				_, err := schema.ParsePostItem(path)
				return err
			},
		},
		{
			name:    "tag title invalid localization",
			content: `{"version":1,"title":{"en":"Title"},"time":1,"author":{"en":"Author"},"tags":[{"title":{"en":1}}]}`,
			parse: func(path string) error {
				_, err := schema.ParsePostItem(path)
				return err
			},
		},
		{
			name:    "playlist levels non string",
			content: `{"version":1,"title":{"en":"Title"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[],"levels":[1]}`,
			parse: func(path string) error {
				_, err := schema.ParsePlaylistItem(path)
				return err
			},
		},
		{
			name:    "engine reference non string",
			content: `{"version":13,"title":{"en":"Title"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[],"skin":1,"background":"background","effect":"effect","particle":"particle"}`,
			parse: func(path string) error {
				_, err := schema.ParseEngineItem(path)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.parse(writeTemp(t, tt.content)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestParseSrlRejectsInvalidShape(t *testing.T) {
	tests := []string{
		`[]`,
		`"bad"`,
		`{"hash":1}`,
	}

	for _, content := range tests {
		t.Run(content, func(t *testing.T) {
			if _, err := schema.ParseSrl(writeTemp(t, content)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestTypedParsersAcceptValidItems(t *testing.T) {
	tests := []struct {
		name    string
		content string
		parse   func(string) error
	}{
		{
			name:    "playlist",
			content: `{"version":1,"title":{"en":"Title"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[],"levels":[]}`,
			parse: func(path string) error {
				_, err := schema.ParsePlaylistItem(path)
				return err
			},
		},
		{
			name:    "skin",
			content: `{"version":4,"title":{"en":"Title"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[]}`,
			parse: func(path string) error {
				_, err := schema.ParseSkinItem(path)
				return err
			},
		},
		{
			name:    "background",
			content: `{"version":2,"title":{"en":"Title"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[]}`,
			parse: func(path string) error {
				_, err := schema.ParseBackgroundItem(path)
				return err
			},
		},
		{
			name:    "effect",
			content: `{"version":5,"title":{"en":"Title"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[]}`,
			parse: func(path string) error {
				_, err := schema.ParseEffectItem(path)
				return err
			},
		},
		{
			name:    "particle",
			content: `{"version":3,"title":{"en":"Title"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[]}`,
			parse: func(path string) error {
				_, err := schema.ParseParticleItem(path)
				return err
			},
		},
		{
			name:    "engine",
			content: `{"version":13,"title":{"en":"Title"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[],"skin":"skin","background":"background","effect":"effect","particle":"particle"}`,
			parse: func(path string) error {
				_, err := schema.ParseEngineItem(path)
				return err
			},
		},
		{
			name:    "replay",
			content: `{"version":1,"title":{"en":"Title"},"subtitle":{"en":"Sub"},"author":{"en":"Author"},"tags":[],"level":"level"}`,
			parse: func(path string) error {
				_, err := schema.ParseReplayItem(path)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.parse(writeTemp(t, tt.content)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "item.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
