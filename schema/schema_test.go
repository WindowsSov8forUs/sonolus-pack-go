package schema_test

import (
	"encoding/json"
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
		"useSkin": { "useDefault": true, "item": "skin" },
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
