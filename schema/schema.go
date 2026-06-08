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
	if fields == nil {
		return fmt.Errorf("%s: must be an object", path)
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
	if err := validateFields(path, fields); err != nil {
		return err
	}

	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func validateFields(path string, fields map[string]json.RawMessage) error {
	for key, raw := range fields {
		var err error
		switch key {
		case "title", "subtitle", "author", "description", "artists":
			err = validateLocalizationText(raw)
		case "tags":
			err = validateTags(raw)
		case "hash", "url":
			err = validateOptionalNullableString(raw)
		case "levels":
			err = validateStringArray(raw)
		case "engine", "skin", "background", "effect", "particle", "level":
			err = validateString(raw)
		case "useSkin", "useBackground", "useEffect", "useParticle":
			err = validateUseItem(raw)
		}
		if err != nil {
			return fmt.Errorf("%s: %s: %w", path, key, err)
		}
	}
	return nil
}

func validateLocalizationText(raw json.RawMessage) error {
	var value map[string]json.RawMessage
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("must be localization text")
	}
	if value == nil {
		return fmt.Errorf("must be localization text")
	}
	for key, rawText := range value {
		if err := validateString(rawText); err != nil {
			return fmt.Errorf("/%s: %w", key, err)
		}
	}
	return nil
}

func validateTags(raw json.RawMessage) error {
	var tags []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &tags); err != nil {
		return fmt.Errorf("must be an array of tags")
	}
	if tags == nil {
		return fmt.Errorf("must be an array of tags")
	}
	for i, tag := range tags {
		if tag == nil {
			return fmt.Errorf("/%d must be an object", i)
		}
		for key, value := range tag {
			switch key {
			case "title":
				if err := validateLocalizationText(value); err != nil {
					return fmt.Errorf("/%d/title: %w", i, err)
				}
			case "icon":
				if err := validateString(value); err != nil {
					return fmt.Errorf("/%d/icon: %w", i, err)
				}
			}
		}
	}
	return nil
}

func validateOptionalNullableString(raw json.RawMessage) error {
	if string(raw) == "null" {
		return nil
	}
	return validateString(raw)
}

func validateStringArray(raw json.RawMessage) error {
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return fmt.Errorf("must be an array of strings")
	}
	if values == nil {
		return fmt.Errorf("must be an array of strings")
	}
	return nil
}

func validateString(raw json.RawMessage) error {
	if string(raw) == "null" {
		return fmt.Errorf("must be a string")
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("must be a string")
	}
	return nil
}

func validateUseItem(raw json.RawMessage) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return fmt.Errorf("must be a use item")
	}
	if fields == nil {
		return fmt.Errorf("must be a use item")
	}

	useDefaultRaw, ok := fields["useDefault"]
	if !ok {
		return fmt.Errorf("useDefault is required")
	}
	var useDefault bool
	if err := json.Unmarshal(useDefaultRaw, &useDefault); err != nil {
		return fmt.Errorf("useDefault must be a boolean")
	}
	if useDefault {
		return nil
	}

	itemRaw, ok := fields["item"]
	if !ok {
		return fmt.Errorf("item is required when useDefault is false")
	}
	if err := validateString(itemRaw); err != nil {
		return fmt.Errorf("item: %w", err)
	}
	return nil
}
