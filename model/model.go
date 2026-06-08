package model

import (
	"encoding/json"
	"fmt"

	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	"github.com/WindowsSov8forUs/sonolus-core-go/database"
)

type Database struct {
	Info        ServerInfo       `json:"info"`
	Posts       []PostItem       `json:"posts"`
	Playlists   []PlaylistItem   `json:"playlists"`
	Levels      []LevelItem      `json:"levels"`
	Skins       []SkinItem       `json:"skins"`
	Backgrounds []BackgroundItem `json:"backgrounds"`
	Effects     []EffectItem     `json:"effects"`
	Particles   []ParticleItem   `json:"particles"`
	Engines     []EngineItem     `json:"engines"`
	Replays     []ReplayItem     `json:"replays"`
}

type ServerInfo struct {
	Title       database.LocalizationText  `json:"title"`
	Description *database.LocalizationText `json:"description,omitempty"`
	Banner      *core.Srl                  `json:"banner,omitempty"`
}

type DatabaseUseItem struct {
	UseDefault bool   `json:"useDefault"`
	Item       string `json:"item,omitempty"`
}

type DatabaseTag struct {
	Title *database.LocalizationText `json:"title,omitempty"`
	Icon  *core.Icon                 `json:"icon,omitempty"`
}

func (u *DatabaseUseItem) UnmarshalJSON(data []byte) error {
	var raw struct {
		UseDefault *bool   `json:"useDefault"`
		Item       *string `json:"item"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.UseDefault == nil {
		return fmt.Errorf("useDefault is required")
	}
	u.UseDefault = *raw.UseDefault
	u.Item = ""
	if !u.UseDefault {
		if raw.Item == nil {
			return fmt.Errorf("item is required when useDefault is false")
		}
		u.Item = *raw.Item
	}
	return nil
}

type PostItem struct {
	Name        string                     `json:"name"`
	Version     int                        `json:"version"`
	Title       database.LocalizationText  `json:"title"`
	Time        float64                    `json:"time"`
	Author      database.LocalizationText  `json:"author"`
	Tags        []DatabaseTag              `json:"tags"`
	Description *database.LocalizationText `json:"description,omitempty"`
	Meta        json.RawMessage            `json:"meta,omitempty"`
	Thumbnail   *core.Srl                  `json:"thumbnail,omitempty"`
}

type PlaylistItem struct {
	Name        string                     `json:"name"`
	Version     int                        `json:"version"`
	Title       database.LocalizationText  `json:"title"`
	Subtitle    database.LocalizationText  `json:"subtitle"`
	Author      database.LocalizationText  `json:"author"`
	Tags        []DatabaseTag              `json:"tags"`
	Description *database.LocalizationText `json:"description,omitempty"`
	Levels      []string                   `json:"levels"`
	Meta        json.RawMessage            `json:"meta,omitempty"`
	Thumbnail   *core.Srl                  `json:"thumbnail,omitempty"`
}

type LevelItem struct {
	Name          string                     `json:"name"`
	Version       int                        `json:"version"`
	Rating        float64                    `json:"rating"`
	Title         database.LocalizationText  `json:"title"`
	Artists       database.LocalizationText  `json:"artists"`
	Author        database.LocalizationText  `json:"author"`
	Tags          []DatabaseTag              `json:"tags"`
	Description   *database.LocalizationText `json:"description,omitempty"`
	Engine        string                     `json:"engine"`
	UseSkin       DatabaseUseItem            `json:"useSkin"`
	UseBackground DatabaseUseItem            `json:"useBackground"`
	UseEffect     DatabaseUseItem            `json:"useEffect"`
	UseParticle   DatabaseUseItem            `json:"useParticle"`
	Meta          json.RawMessage            `json:"meta,omitempty"`
	Cover         core.Srl                   `json:"cover"`
	BGM           core.Srl                   `json:"bgm"`
	Preview       *core.Srl                  `json:"preview,omitempty"`
	Data          core.Srl                   `json:"data"`
}

type SkinItem struct {
	Name        string                     `json:"name"`
	Version     int                        `json:"version"`
	Title       database.LocalizationText  `json:"title"`
	Subtitle    database.LocalizationText  `json:"subtitle"`
	Author      database.LocalizationText  `json:"author"`
	Tags        []DatabaseTag              `json:"tags"`
	Description *database.LocalizationText `json:"description,omitempty"`
	Meta        json.RawMessage            `json:"meta,omitempty"`
	Thumbnail   core.Srl                   `json:"thumbnail"`
	Data        core.Srl                   `json:"data"`
	Texture     core.Srl                   `json:"texture"`
}

type BackgroundItem struct {
	Name          string                     `json:"name"`
	Version       int                        `json:"version"`
	Title         database.LocalizationText  `json:"title"`
	Subtitle      database.LocalizationText  `json:"subtitle"`
	Author        database.LocalizationText  `json:"author"`
	Tags          []DatabaseTag              `json:"tags"`
	Description   *database.LocalizationText `json:"description,omitempty"`
	Meta          json.RawMessage            `json:"meta,omitempty"`
	Thumbnail     core.Srl                   `json:"thumbnail"`
	Data          core.Srl                   `json:"data"`
	Image         core.Srl                   `json:"image"`
	Configuration core.Srl                   `json:"configuration"`
}

type EffectItem struct {
	Name        string                     `json:"name"`
	Version     int                        `json:"version"`
	Title       database.LocalizationText  `json:"title"`
	Subtitle    database.LocalizationText  `json:"subtitle"`
	Author      database.LocalizationText  `json:"author"`
	Tags        []DatabaseTag              `json:"tags"`
	Description *database.LocalizationText `json:"description,omitempty"`
	Meta        json.RawMessage            `json:"meta,omitempty"`
	Thumbnail   core.Srl                   `json:"thumbnail"`
	Data        core.Srl                   `json:"data"`
	Audio       core.Srl                   `json:"audio"`
}

type ParticleItem struct {
	Name        string                     `json:"name"`
	Version     int                        `json:"version"`
	Title       database.LocalizationText  `json:"title"`
	Subtitle    database.LocalizationText  `json:"subtitle"`
	Author      database.LocalizationText  `json:"author"`
	Tags        []DatabaseTag              `json:"tags"`
	Description *database.LocalizationText `json:"description,omitempty"`
	Meta        json.RawMessage            `json:"meta,omitempty"`
	Thumbnail   core.Srl                   `json:"thumbnail"`
	Data        core.Srl                   `json:"data"`
	Texture     core.Srl                   `json:"texture"`
}

type EngineItem struct {
	Name          string                     `json:"name"`
	Version       int                        `json:"version"`
	Title         database.LocalizationText  `json:"title"`
	Subtitle      database.LocalizationText  `json:"subtitle"`
	Author        database.LocalizationText  `json:"author"`
	Tags          []DatabaseTag              `json:"tags"`
	Description   *database.LocalizationText `json:"description,omitempty"`
	Skin          string                     `json:"skin"`
	Background    string                     `json:"background"`
	Effect        string                     `json:"effect"`
	Particle      string                     `json:"particle"`
	Meta          json.RawMessage            `json:"meta,omitempty"`
	Thumbnail     core.Srl                   `json:"thumbnail"`
	PlayData      core.Srl                   `json:"playData"`
	WatchData     core.Srl                   `json:"watchData"`
	PreviewData   core.Srl                   `json:"previewData"`
	TutorialData  core.Srl                   `json:"tutorialData"`
	ROM           *core.Srl                  `json:"rom,omitempty"`
	Configuration core.Srl                   `json:"configuration"`
}

type ReplayItem struct {
	Name          string                     `json:"name"`
	Version       int                        `json:"version"`
	Title         database.LocalizationText  `json:"title"`
	Subtitle      database.LocalizationText  `json:"subtitle"`
	Author        database.LocalizationText  `json:"author"`
	Tags          []DatabaseTag              `json:"tags"`
	Description   *database.LocalizationText `json:"description,omitempty"`
	Level         string                     `json:"level"`
	Meta          json.RawMessage            `json:"meta,omitempty"`
	Data          core.Srl                   `json:"data"`
	Configuration core.Srl                   `json:"configuration"`
}
