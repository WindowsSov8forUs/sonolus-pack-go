package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	"github.com/WindowsSov8forUs/sonolus-pack-go/model"
)

type ValidationError struct {
	Path     string
	JSONPath string
	Message  string
	Value    json.RawMessage
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Invalid data: %s", e.Path)
}

type ValidationErrors struct {
	Path  string
	Items []ValidationError
}

func (e *ValidationErrors) Error() string {
	return fmt.Sprintf("Invalid data: %s", e.Path)
}

func ParseInfo(path string) (model.ServerInfo, error) {
	var item model.ServerInfo
	if err := parse(path, &item, []string{"title", "description"}, []string{"title"}, 0); err != nil {
		return model.ServerInfo{}, err
	}
	return item, nil
}

func ParseSrl(path string) (core.Srl, error) {
	var srl core.Srl
	if err := parse(path, &srl, []string{"hash", "url"}, nil, 0); err != nil {
		return core.Srl{}, err
	}
	return srl, nil
}

func ParsePostItem(path string) (model.PostItem, error) {
	var item model.PostItem
	return item, parse(path, &item, []string{"version", "title", "time", "author", "tags", "description", "meta"}, []string{"version", "title", "time", "author", "tags"}, 1)
}

func ParsePlaylistItem(path string) (model.PlaylistItem, error) {
	var item model.PlaylistItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags", "description", "levels", "meta"}, []string{"version", "title", "subtitle", "author", "tags", "levels"}, 1)
}

func ParseLevelItem(path string) (model.LevelItem, error) {
	var item model.LevelItem
	return item, parse(path, &item, []string{"version", "rating", "title", "artists", "author", "tags", "description", "engine", "useSkin", "useBackground", "useEffect", "useParticle", "meta"}, []string{"version", "rating", "title", "artists", "author", "tags", "engine", "useSkin", "useBackground", "useEffect", "useParticle"}, 1)
}

func ParseSkinItem(path string) (model.SkinItem, error) {
	var item model.SkinItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags", "description", "meta"}, []string{"version", "title", "subtitle", "author", "tags"}, 4)
}

func ParseBackgroundItem(path string) (model.BackgroundItem, error) {
	var item model.BackgroundItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags", "description", "meta"}, []string{"version", "title", "subtitle", "author", "tags"}, 2)
}

func ParseEffectItem(path string) (model.EffectItem, error) {
	var item model.EffectItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags", "description", "meta"}, []string{"version", "title", "subtitle", "author", "tags"}, 5)
}

func ParseParticleItem(path string) (model.ParticleItem, error) {
	var item model.ParticleItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags", "description", "meta"}, []string{"version", "title", "subtitle", "author", "tags"}, 3)
}

func ParseEngineItem(path string) (model.EngineItem, error) {
	var item model.EngineItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags", "description", "skin", "background", "effect", "particle", "meta"}, []string{"version", "title", "subtitle", "author", "tags", "skin", "background", "effect", "particle"}, 13)
}

func ParseReplayItem(path string) (model.ReplayItem, error) {
	var item model.ReplayItem
	return item, parse(path, &item, []string{"version", "title", "subtitle", "author", "tags", "description", "level", "meta"}, []string{"version", "title", "subtitle", "author", "tags", "level"}, 1)
}

func parse(path string, out any, allowed, required []string, version int) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: Does not exist", displayPath(path))
		}
		return err
	}

	if err := validateTopLevelObject(path, data); err != nil {
		return err
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return fmt.Errorf("%s: %w", displayPath(path), err)
	}

	var validationErrors []ValidationError
	missing := map[string]bool{}
	for _, key := range required {
		raw, ok := fields[key]
		if !ok {
			missing[key] = true
			validationErrors = append(validationErrors, newValidationError(path, "/"+key, key+" is required", nil))
			continue
		}
		if len(raw) == 0 {
			validationErrors = append(validationErrors, newValidationError(path, "/"+key, key+" is required", raw))
		}
	}
	if version != 0 && !missing["version"] {
		if validationErr := validateVersion(path, fields, version); validationErr != nil {
			validationErrors = append(validationErrors, *validationErr)
		}
	}
	validationErrors = append(validationErrors, validateFields(path, fields, allowed)...)
	if len(validationErrors) != 0 {
		return &ValidationErrors{Path: displayPath(path), Items: validationErrors}
	}

	if data, err = json.Marshal(fields); err != nil {
		return fmt.Errorf("%s: %w", displayPath(path), err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("%s: %w", displayPath(path), err)
	}
	return nil
}

func validateTopLevelObject(path string, data []byte) error {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("%s: %w", displayPath(path), err)
	}
	if _, ok := value.(map[string]any); !ok {
		validationErr := newValidationError(path, "", "must be an object", data)
		return &ValidationErrors{Path: displayPath(path), Items: []ValidationError{validationErr}}
	}
	return nil
}

func validateFields(path string, fields map[string]json.RawMessage, allowed []string) []ValidationError {
	var errors []ValidationError
	for _, key := range allowed {
		raw, ok := fields[key]
		if !ok {
			continue
		}
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
		case "time", "rating":
			err = validateNumber(raw)
		}
		if err != nil {
			errors = append(errors, newValidationError(path, "/"+key, fmt.Sprintf("%s: %s", key, err), raw))
		}
	}
	return errors
}

func newValidationError(path, jsonPath, message string, value json.RawMessage) ValidationError {
	return ValidationError{
		Path:     displayPath(path),
		JSONPath: jsonPath,
		Message:  message,
		Value:    value,
	}
}

func displayPath(path string) string {
	return filepath.ToSlash(path)
}

func validateVersion(path string, fields map[string]json.RawMessage, version int) *ValidationError {
	raw := fields["version"]
	if string(raw) == "null" {
		validationErr := newValidationError(path, "/version", fmt.Sprintf("version must be %d", version), raw)
		return &validationErr
	}
	var got float64
	if err := json.Unmarshal(raw, &got); err != nil {
		validationErr := newValidationError(path, "/version", fmt.Sprintf("version must be %d", version), raw)
		return &validationErr
	}
	if got != float64(version) {
		validationErr := newValidationError(path, "/version", fmt.Sprintf("version must be %d, got %s", version, string(raw)), raw)
		return &validationErr
	}
	fields["version"] = json.RawMessage(strconv.Itoa(version))
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

func validateNumber(raw json.RawMessage) error {
	if string(raw) == "null" {
		return fmt.Errorf("must be a number")
	}
	var value float64
	if err := json.Unmarshal(raw, &value); err != nil {
		return fmt.Errorf("must be a number")
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
