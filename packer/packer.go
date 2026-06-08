package packer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	"github.com/WindowsSov8forUs/sonolus-pack-go/model"
	"github.com/WindowsSov8forUs/sonolus-pack-go/resource"
	"github.com/WindowsSov8forUs/sonolus-pack-go/schema"
)

type Logger interface {
	Info(args ...any)
	Warning(args ...any)
}

type Options struct {
	Input  string
	Output string
	Logger Logger
}

func Pack(ctx context.Context, options Options) error {
	if options.Input == "" {
		options.Input = "source"
	}
	if options.Output == "" {
		options.Output = "pack"
	}
	if err := os.RemoveAll(options.Output); err != nil {
		return err
	}
	if err := os.MkdirAll(options.Output, 0o755); err != nil {
		return err
	}

	db, err := build(ctx, options)
	if err != nil {
		_ = os.RemoveAll(options.Output)
		return err
	}

	data, err := json.Marshal(db)
	if err != nil {
		_ = os.RemoveAll(options.Output)
		return err
	}
	if err := os.WriteFile(filepath.Join(options.Output, "db.json"), data, 0o644); err != nil {
		_ = os.RemoveAll(options.Output)
		return err
	}
	return nil
}

func build(ctx context.Context, options Options) (model.Database, error) {
	packer := resource.Packer{
		Output: options.Output,
		Logger: options.Logger,
	}

	infoDir := options.Input
	logPacking(packer.Logger, infoDir)
	info, err := schema.ParseInfo(filepath.Join(infoDir, "info.json"))
	if err != nil {
		return model.Database{}, err
	}
	if srl, ok, err := packer.Pack(filepath.Join(options.Input, "banner"), "png", true); err != nil {
		return model.Database{}, err
	} else if ok {
		info.Banner = srl
	}

	db := model.Database{Info: info}
	if db.Posts, err = processPosts(ctx, options.Input, packer); err != nil {
		return model.Database{}, err
	}
	if db.Playlists, err = processPlaylists(ctx, options.Input, packer); err != nil {
		return model.Database{}, err
	}
	if db.Levels, err = processLevels(ctx, options.Input, packer); err != nil {
		return model.Database{}, err
	}
	if db.Skins, err = processSkins(ctx, options.Input, packer); err != nil {
		return model.Database{}, err
	}
	if db.Backgrounds, err = processBackgrounds(ctx, options.Input, packer); err != nil {
		return model.Database{}, err
	}
	if db.Effects, err = processEffects(ctx, options.Input, packer); err != nil {
		return model.Database{}, err
	}
	if db.Particles, err = processParticles(ctx, options.Input, packer); err != nil {
		return model.Database{}, err
	}
	if db.Engines, err = processEngines(ctx, options.Input, packer); err != nil {
		return model.Database{}, err
	}
	if db.Replays, err = processReplays(ctx, options.Input, packer); err != nil {
		return model.Database{}, err
	}

	if err := checkReferences(db); err != nil {
		return model.Database{}, err
	}
	return db, nil
}

func processPosts(ctx context.Context, input string, packer resource.Packer) ([]model.PostItem, error) {
	var items []model.PostItem
	err := eachItemDir(ctx, filepath.Join(input, "posts"), func(name, dir string) error {
		logPacking(packer.Logger, dir)
		item, err := schema.ParsePostItem(filepath.Join(dir, "item.json"))
		if err != nil {
			return err
		}
		item.Name = name
		if srl, ok, err := packer.Pack(filepath.Join(dir, "thumbnail"), "png", true); err != nil {
			return err
		} else if ok {
			item.Thumbnail = srl
		}
		items = append(items, item)
		return nil
	})
	return items, err
}

func processPlaylists(ctx context.Context, input string, packer resource.Packer) ([]model.PlaylistItem, error) {
	var items []model.PlaylistItem
	err := eachItemDir(ctx, filepath.Join(input, "playlists"), func(name, dir string) error {
		logPacking(packer.Logger, dir)
		item, err := schema.ParsePlaylistItem(filepath.Join(dir, "item.json"))
		if err != nil {
			return err
		}
		item.Name = name
		if srl, ok, err := packer.Pack(filepath.Join(dir, "thumbnail"), "png", true); err != nil {
			return err
		} else if ok {
			item.Thumbnail = srl
		}
		items = append(items, item)
		return nil
	})
	return items, err
}

func processLevels(ctx context.Context, input string, packer resource.Packer) ([]model.LevelItem, error) {
	var items []model.LevelItem
	err := eachItemDir(ctx, filepath.Join(input, "levels"), func(name, dir string) error {
		logPacking(packer.Logger, dir)
		item, err := schema.ParseLevelItem(filepath.Join(dir, "item.json"))
		if err != nil {
			return err
		}
		item.Name = name
		if item.Cover, err = required(packer.Pack(filepath.Join(dir, "cover"), "png", false)); err != nil {
			return err
		}
		if item.BGM, err = required(packer.Pack(filepath.Join(dir, "bgm"), "mp3", false)); err != nil {
			return err
		}
		if srl, ok, err := packer.Pack(filepath.Join(dir, "preview"), "mp3", true); err != nil {
			return err
		} else if ok {
			item.Preview = srl
		}
		if item.Data, err = required(packer.Pack(filepath.Join(dir, "data"), "json", false)); err != nil {
			return err
		}
		items = append(items, item)
		return nil
	})
	return items, err
}

func processSkins(ctx context.Context, input string, packer resource.Packer) ([]model.SkinItem, error) {
	var items []model.SkinItem
	err := eachItemDir(ctx, filepath.Join(input, "skins"), func(name, dir string) error {
		logPacking(packer.Logger, dir)
		item, err := schema.ParseSkinItem(filepath.Join(dir, "item.json"))
		if err != nil {
			return err
		}
		item.Name = name
		if item.Thumbnail, err = required(packer.Pack(filepath.Join(dir, "thumbnail"), "png", false)); err != nil {
			return err
		}
		if item.Data, err = required(packer.Pack(filepath.Join(dir, "data"), "json", false)); err != nil {
			return err
		}
		if item.Texture, err = required(packer.Pack(filepath.Join(dir, "texture"), "png", false)); err != nil {
			return err
		}
		items = append(items, item)
		return nil
	})
	return items, err
}

func processBackgrounds(ctx context.Context, input string, packer resource.Packer) ([]model.BackgroundItem, error) {
	var items []model.BackgroundItem
	err := eachItemDir(ctx, filepath.Join(input, "backgrounds"), func(name, dir string) error {
		logPacking(packer.Logger, dir)
		item, err := schema.ParseBackgroundItem(filepath.Join(dir, "item.json"))
		if err != nil {
			return err
		}
		item.Name = name
		if item.Thumbnail, err = required(packer.Pack(filepath.Join(dir, "thumbnail"), "png", false)); err != nil {
			return err
		}
		if item.Data, err = required(packer.Pack(filepath.Join(dir, "data"), "json", false)); err != nil {
			return err
		}
		if item.Image, err = required(packer.Pack(filepath.Join(dir, "image"), "png", false)); err != nil {
			return err
		}
		if item.Configuration, err = required(packer.Pack(filepath.Join(dir, "configuration"), "json", false)); err != nil {
			return err
		}
		items = append(items, item)
		return nil
	})
	return items, err
}

func processEffects(ctx context.Context, input string, packer resource.Packer) ([]model.EffectItem, error) {
	var items []model.EffectItem
	err := eachItemDir(ctx, filepath.Join(input, "effects"), func(name, dir string) error {
		logPacking(packer.Logger, dir)
		item, err := schema.ParseEffectItem(filepath.Join(dir, "item.json"))
		if err != nil {
			return err
		}
		item.Name = name
		if item.Thumbnail, err = required(packer.Pack(filepath.Join(dir, "thumbnail"), "png", false)); err != nil {
			return err
		}
		if item.Data, err = required(packer.Pack(filepath.Join(dir, "data"), "json", false)); err != nil {
			return err
		}
		if item.Audio, err = required(packer.Pack(filepath.Join(dir, "audio"), "zip", false)); err != nil {
			return err
		}
		items = append(items, item)
		return nil
	})
	return items, err
}

func processParticles(ctx context.Context, input string, packer resource.Packer) ([]model.ParticleItem, error) {
	var items []model.ParticleItem
	err := eachItemDir(ctx, filepath.Join(input, "particles"), func(name, dir string) error {
		logPacking(packer.Logger, dir)
		item, err := schema.ParseParticleItem(filepath.Join(dir, "item.json"))
		if err != nil {
			return err
		}
		item.Name = name
		if item.Thumbnail, err = required(packer.Pack(filepath.Join(dir, "thumbnail"), "png", false)); err != nil {
			return err
		}
		if item.Data, err = required(packer.Pack(filepath.Join(dir, "data"), "json", false)); err != nil {
			return err
		}
		if item.Texture, err = required(packer.Pack(filepath.Join(dir, "texture"), "png", false)); err != nil {
			return err
		}
		items = append(items, item)
		return nil
	})
	return items, err
}

func processEngines(ctx context.Context, input string, packer resource.Packer) ([]model.EngineItem, error) {
	var items []model.EngineItem
	err := eachItemDir(ctx, filepath.Join(input, "engines"), func(name, dir string) error {
		logPacking(packer.Logger, dir)
		item, err := schema.ParseEngineItem(filepath.Join(dir, "item.json"))
		if err != nil {
			return err
		}
		item.Name = name
		if item.Thumbnail, err = required(packer.Pack(filepath.Join(dir, "thumbnail"), "png", false)); err != nil {
			return err
		}
		if item.PlayData, err = required(packer.Pack(filepath.Join(dir, "playData"), "json", false)); err != nil {
			return err
		}
		if item.WatchData, err = required(packer.Pack(filepath.Join(dir, "watchData"), "json", false)); err != nil {
			return err
		}
		if item.PreviewData, err = required(packer.Pack(filepath.Join(dir, "previewData"), "json", false)); err != nil {
			return err
		}
		if item.TutorialData, err = required(packer.Pack(filepath.Join(dir, "tutorialData"), "json", false)); err != nil {
			return err
		}
		if srl, ok, err := packer.Pack(filepath.Join(dir, "rom"), "bin", true); err != nil {
			return err
		} else if ok {
			item.ROM = srl
		}
		if item.Configuration, err = required(packer.Pack(filepath.Join(dir, "configuration"), "json", false)); err != nil {
			return err
		}
		items = append(items, item)
		return nil
	})
	return items, err
}

func processReplays(ctx context.Context, input string, packer resource.Packer) ([]model.ReplayItem, error) {
	var items []model.ReplayItem
	err := eachItemDir(ctx, filepath.Join(input, "replays"), func(name, dir string) error {
		logPacking(packer.Logger, dir)
		item, err := schema.ParseReplayItem(filepath.Join(dir, "item.json"))
		if err != nil {
			return err
		}
		item.Name = name
		if item.Data, err = required(packer.Pack(filepath.Join(dir, "data"), "json", false)); err != nil {
			return err
		}
		if item.Configuration, err = required(packer.Pack(filepath.Join(dir, "configuration"), "json", false)); err != nil {
			return err
		}
		items = append(items, item)
		return nil
	})
	return items, err
}

func logPacking(logger Logger, path string) {
	if logger != nil {
		logger.Info("Packing:", path)
	}
}

func eachItemDir(ctx context.Context, dir string, fn func(name, dir string) error) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if err := fn(name, filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

func required(srl *core.Srl, ok bool, err error) (core.Srl, error) {
	if err != nil {
		return core.Srl{}, err
	}
	if !ok || srl == nil {
		return core.Srl{}, nil
	}
	return *srl, nil
}

func checkReferences(db model.Database) error {
	levelNames := names(db.Levels, func(item model.LevelItem) string { return item.Name })
	skinNames := names(db.Skins, func(item model.SkinItem) string { return item.Name })
	backgroundNames := names(db.Backgrounds, func(item model.BackgroundItem) string { return item.Name })
	effectNames := names(db.Effects, func(item model.EffectItem) string { return item.Name })
	particleNames := names(db.Particles, func(item model.ParticleItem) string { return item.Name })
	engineNames := names(db.Engines, func(item model.EngineItem) string { return item.Name })

	for _, playlist := range db.Playlists {
		parent := "playlists/" + playlist.Name
		for i, level := range playlist.Levels {
			if !levelNames[level] {
				return fmt.Errorf("%s: %s not found (/levels/%d)", parent, level, i)
			}
		}
	}
	for _, level := range db.Levels {
		parent := "levels/" + level.Name
		if !engineNames[level.Engine] {
			return fmt.Errorf("%s: %s not found (/engine)", parent, level.Engine)
		}
		if !level.UseSkin.UseDefault && !skinNames[level.UseSkin.Item] {
			return fmt.Errorf("%s: %s not found (/useSkin/item)", parent, level.UseSkin.Item)
		}
		if !level.UseBackground.UseDefault && !backgroundNames[level.UseBackground.Item] {
			return fmt.Errorf("%s: %s not found (/useBackground/item)", parent, level.UseBackground.Item)
		}
		if !level.UseEffect.UseDefault && !effectNames[level.UseEffect.Item] {
			return fmt.Errorf("%s: %s not found (/useEffect/item)", parent, level.UseEffect.Item)
		}
		if !level.UseParticle.UseDefault && !particleNames[level.UseParticle.Item] {
			return fmt.Errorf("%s: %s not found (/useParticle/item)", parent, level.UseParticle.Item)
		}
	}
	for _, engine := range db.Engines {
		parent := "engines/" + engine.Name
		if !skinNames[engine.Skin] {
			return fmt.Errorf("%s: %s not found (/skin)", parent, engine.Skin)
		}
		if !backgroundNames[engine.Background] {
			return fmt.Errorf("%s: %s not found (/background)", parent, engine.Background)
		}
		if !effectNames[engine.Effect] {
			return fmt.Errorf("%s: %s not found (/effect)", parent, engine.Effect)
		}
		if !particleNames[engine.Particle] {
			return fmt.Errorf("%s: %s not found (/particle)", parent, engine.Particle)
		}
	}
	for _, replay := range db.Replays {
		parent := "replays/" + replay.Name
		if !levelNames[replay.Level] {
			return fmt.Errorf("%s: %s not found (/level)", parent, replay.Level)
		}
	}
	return nil
}

func names[T any](items []T, name func(T) string) map[string]bool {
	result := map[string]bool{}
	for _, item := range items {
		result[name(item)] = true
	}
	return result
}
