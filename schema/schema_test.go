package schema_test

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestParseSrlAcceptsNullableFields(t *testing.T) {
	path := writeTemp(t, `{"hash": null, "url": "https://example.com"}`)

	if _, err := schema.ParseSrl(path); err != nil {
		t.Fatal(err)
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
