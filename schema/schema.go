package schema

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	"github.com/WindowsSov8forUs/sonolus-pack-go/model"
)

func ParseInfo(path string) (model.ServerInfo, error) {
	var item model.ServerInfo
	if err := parse(path, &item, []string{"title"}, 0); err != nil {
		return model.ServerInfo{}, err
	}
	return item, nil
}

func ParseSrl(path string) (core.Srl, error) {
	var srl core.Srl
	if err := parse(path, &srl, nil, 0); err != nil {
		return core.Srl{}, err
	}
	return srl, nil
}

func ParsePostItem(path string) (model.PostItem, error) {
	var item model.PostItem
	return item, parse(path, &item, []string{"version", "title", "time", "author", "tags"}, 1)
}

func ParsePlaylistItem(path string) (model.PlaylistItem, error) {
	var item model.PlaylistItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags", "levels"}, 1)
}

func ParseLevelItem(path string) (model.LevelItem, error) {
	var item model.LevelItem
	return item, parse(path, &item, []string{"version", "rating", "title", "artists", "author", "tags", "engine", "useSkin", "useBackground", "useEffect", "useParticle"}, 1)
}

func ParseSkinItem(path string) (model.SkinItem, error) {
	var item model.SkinItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags"}, 4)
}

func ParseBackgroundItem(path string) (model.BackgroundItem, error) {
	var item model.BackgroundItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags"}, 2)
}

func ParseEffectItem(path string) (model.EffectItem, error) {
	var item model.EffectItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags"}, 5)
}

func ParseParticleItem(path string) (model.ParticleItem, error) {
	var item model.ParticleItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags"}, 3)
}

func ParseEngineItem(path string) (model.EngineItem, error) {
	var item model.EngineItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags", "skin", "background", "effect", "particle"}, 13)
}

func ParseReplayItem(path string) (model.ReplayItem, error) {
	var item model.ReplayItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags", "level"}, 1)
}

func parse(path string, out any, required []string, version int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	for _, key := range required {
		if _, ok := fields[key]; !ok {
			return fmt.Errorf("%s: %s is required", path, key)
		}
	}
	if version != 0 {
		var got int
		if err := json.Unmarshal(fields["version"], &got); err != nil {
			return fmt.Errorf("%s: version must be %d", path, version)
		}
		if got != version {
			return fmt.Errorf("%s: version must be %d, got %d", path, version, got)
		}
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
