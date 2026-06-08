# Sonolus Pack

用于将 Sonolus source 文件打包为 repository 与数据库的命令行工具。

这是 `@sonolus/pack` 的 Go 逻辑复刻实现。主要交付物是可直接下载运行的 `sonolus-pack` 二进制程序，不需要安装 Go、Node.js 或 npm。

## 链接

- [Sonolus 官网](https://sonolus.com)
- [Sonolus Wiki](https://wiki.sonolus.com)
- [sonolus-express](https://github.com/Sonolus/sonolus-express)
- [sonolus-generate-static](https://github.com/Sonolus/sonolus-generate-static)

## 使用

从 GitHub Release 下载平台对应的压缩包，解压后运行：

```powershell
.\sonolus-pack.exe -i source -o pack
```

Linux/macOS：

```sh
./sonolus-pack -i source -o pack
```

使用默认选项打包：

```sh
sonolus-pack
```

查看可用选项：

```sh
sonolus-pack -h
```

查看版本：

```sh
sonolus-pack --version
```

也可以使用短参数：

```sh
sonolus-pack -V
```

参数：

- `-i`、`--input`：输入 source 目录，默认为 `source`
- `-o`、`--output`：输出 pack 目录，默认为 `pack`
- `-V`、`--version`：输出版本号

也可以在 Go 项目中直接调用库入口：

```go
package main

import (
	"context"

	"github.com/WindowsSov8forUs/sonolus-pack-go/pack"
)

func main() {
	err := pack.Pack(context.Background(), pack.Options{
		Input:  "source",
		Output: "pack",
	})
	if err != nil {
		panic(err)
	}
}
```

## 输入

输入目录默认为 `source`。

每个条目的目录名会作为该条目的 `name` 写入数据库。每个 `item.json` 会按 pack schema 清理：保留已声明字段和 `meta`，忽略未知字段。

### 资源文件

资源文件指除 `item.json` 外的各类文件。

每个资源会按以下优先级解析：

1. `name.srl`
2. `name`
3. `name.<默认扩展名>`

如果提供 `.srl` 文件，会直接使用其中的 SRL 对象。

如果提供无扩展名文件，会跳过默认扩展名对应的处理，并为原文件生成 SRL。

如果提供默认扩展名文件，会按资源类型处理并生成 SRL：

- `.json`：使用 Sonolus codec 压缩后写入 repository
- `.bin`：使用 gzip 最高压缩等级压缩后写入 repository
- 其他文件：原样写入 repository

例如资源 `cover[.png/.srl]`：

- 如果提供 `cover.srl`，会直接使用其中的 SRL 对象。
- 如果提供 `cover`，会跳过 `.png` 处理并原样写入 repository，然后生成对应 SRL。
- 如果提供 `cover.png`，会按 `.png` 资源处理并生成对应 SRL。

可选资源缺失时会跳过。必填资源缺失时会记录 warning，并在数据库中写入空 SRL 对象。

本项目目标是逻辑复刻 `@sonolus/pack`，不承诺压缩字节、repository hash 或 JSON key 顺序与 TypeScript 版本完全一致。

### 本地化文本

本地化文本是以语言代码为 key 的对象。

```go
type LocalizationText map[string]string
```

示例：

```json
{
	"en": "Hello!",
	"zhs": "你好！",
	"ja": "こんにちは！",
	"ko": "안녕하세요!"
}
```

### 标签

标签的通用类型：

```go
type Tag struct {
	Title LocalizationText `json:"title,omitempty"`
	Icon  string           `json:"icon,omitempty"`
}
```

示例：

```json
{
	"title": {
		"en": "Tag"
	},
	"icon": "star"
}
```

### 目录结构

```txt
source/
  info.json
  banner[.png/.srl]
  posts/
    {post}/
      item.json
      thumbnail[.png/.srl]
  playlists/
    {playlist}/
      item.json
      thumbnail[.png/.srl]
  levels/
    {level}/
      item.json
      cover[.png/.srl]
      bgm[.mp3/.srl]
      preview[.mp3/.srl]
      data[.json/.srl]
  skins/
    {skin}/
      item.json
      thumbnail[.png/.srl]
      data[.json/.srl]
      texture[.png/.srl]
  backgrounds/
    {background}/
      item.json
      thumbnail[.png/.srl]
      data[.json/.srl]
      image[.png/.srl]
      configuration[.json/.srl]
  effects/
    {effect}/
      item.json
      thumbnail[.png/.srl]
      data[.json/.srl]
      audio[.zip/.srl]
  particles/
    {particle}/
      item.json
      thumbnail[.png/.srl]
      data[.json/.srl]
      texture[.png/.srl]
  engines/
    {engine}/
      item.json
      thumbnail[.png/.srl]
      playData[.json/.srl]
      watchData[.json/.srl]
      previewData[.json/.srl]
      tutorialData[.json/.srl]
      rom[.bin/.srl]
      configuration[.json/.srl]
  replays/
    {replay}/
      item.json
      data[.json/.srl]
      configuration[.json/.srl]
```

### Server

服务器资源位于 source 根目录。

#### `info.json`

服务器信息。

```go
type ServerInfo struct {
	Title       database.LocalizationText `json:"title"`
	Description database.LocalizationText `json:"description,omitempty"`
}
```

示例：

```json
{
	"title": {
		"en": "Server Title"
	},
	"description": {
		"en": "Server Description"
	}
}
```

字段：

- `title`：必填，本地化文本
- `description`：可选，本地化文本

#### `banner[.png/.srl]`

可选，服务器横幅图。

### Posts

每个 post 的资源位于 `posts/{post}`。

#### `item.json`

Post 信息。

```go
type PostItem struct {
	Version     int                       `json:"version"`
	Title       database.LocalizationText `json:"title"`
	Time        float64                   `json:"time"`
	Author      database.LocalizationText `json:"author"`
	Tags        []database.DatabaseTag    `json:"tags"`
	Description database.LocalizationText `json:"description,omitempty"`
	Meta        json.RawMessage           `json:"meta,omitempty"`
}
```

示例：

```json
{
	"version": 1,
	"title": {
		"en": "Post Title"
	},
	"time": 1710000000,
	"author": {
		"en": "Author"
	},
	"tags": [],
	"description": {
		"en": "Post Description"
	},
	"meta": {
		"custom": true
	}
}
```

必填字段：`version`、`title`、`time`、`author`、`tags`。

`version` 必须为 `1`。

#### `thumbnail[.png/.srl]`

可选，post 缩略图。

### Playlists

每个 playlist 的资源位于 `playlists/{playlist}`。

#### `item.json`

Playlist 信息。

```go
type PlaylistItem struct {
	Version     int                       `json:"version"`
	Title       database.LocalizationText `json:"title"`
	Subtitle    database.LocalizationText `json:"subtitle"`
	Author      database.LocalizationText `json:"author"`
	Tags        []database.DatabaseTag    `json:"tags"`
	Description database.LocalizationText `json:"description,omitempty"`
	Levels      []string                  `json:"levels"`
	Meta        json.RawMessage           `json:"meta,omitempty"`
}
```

示例：

```json
{
	"version": 1,
	"title": {
		"en": "Playlist Title"
	},
	"subtitle": {
		"en": "Playlist Subtitle"
	},
	"author": {
		"en": "Author"
	},
	"tags": [],
	"description": {
		"en": "Playlist Description"
	},
	"levels": ["level-name"],
	"meta": {
		"custom": true
	}
}
```

必填字段：`version`、`title`、`subtitle`、`author`、`tags`、`levels`。

`version` 必须为 `1`。`levels` 中的每个名称都必须能在 `levels/` 中找到。

#### `thumbnail[.png/.srl]`

可选，playlist 缩略图。

### Levels

每个 level 的资源位于 `levels/{level}`。

#### `item.json`

Level 信息。

```go
type LevelItem struct {
	Version       int                       `json:"version"`
	Rating        float64                   `json:"rating"`
	Title         database.LocalizationText `json:"title"`
	Artists       database.LocalizationText `json:"artists"`
	Author        database.LocalizationText `json:"author"`
	Tags          []database.DatabaseTag    `json:"tags"`
	Description   database.LocalizationText `json:"description,omitempty"`
	Engine        string                    `json:"engine"`
	UseSkin       DatabaseUseItem           `json:"useSkin"`
	UseBackground DatabaseUseItem           `json:"useBackground"`
	UseEffect     DatabaseUseItem           `json:"useEffect"`
	UseParticle   DatabaseUseItem           `json:"useParticle"`
	Meta          json.RawMessage           `json:"meta,omitempty"`
}

type DatabaseUseItem struct {
	UseDefault bool   `json:"useDefault"`
	Item       string `json:"item,omitempty"`
}
```

示例：

```json
{
	"version": 1,
	"rating": 30,
	"title": {
		"en": "Level Title"
	},
	"artists": {
		"en": "Artist"
	},
	"author": {
		"en": "Author"
	},
	"tags": [],
	"engine": "engine-name",
	"useSkin": {
		"useDefault": true
	},
	"useBackground": {
		"useDefault": true
	},
	"useEffect": {
		"useDefault": true
	},
	"useParticle": {
		"useDefault": true
	},
	"meta": {
		"custom": true
	}
}
```

必填字段：`version`、`rating`、`title`、`artists`、`author`、`tags`、`engine`、`useSkin`、`useBackground`、`useEffect`、`useParticle`。

`version` 必须为 `1`。`engine` 必须能在 `engines/` 中找到。

`useSkin`、`useBackground`、`useEffect`、`useParticle` 可以使用默认资源：

```json
{
	"useDefault": true
}
```

也可以引用指定条目：

```json
{
	"useDefault": false,
	"item": "skin-name"
}
```

当 `useDefault` 为 `false` 时，`item` 必填，并且必须能在对应目录中找到。

#### `cover[.png/.srl]`

Level 封面。

#### `bgm[.mp3/.srl]`

Level 音乐。

#### `preview[.mp3/.srl]`

可选，level 预览音频。

#### `data[.json/.srl]`

Level 数据。

### Skins

每个 skin 的资源位于 `skins/{skin}`。

#### `item.json`

Skin 信息。

```go
type SkinItem struct {
	Version     int                       `json:"version"`
	Title       database.LocalizationText `json:"title"`
	Subtitle    database.LocalizationText `json:"subtitle"`
	Author      database.LocalizationText `json:"author"`
	Tags        []database.DatabaseTag    `json:"tags"`
	Description database.LocalizationText `json:"description,omitempty"`
	Meta        json.RawMessage           `json:"meta,omitempty"`
}
```

示例：

```json
{
	"version": 4,
	"title": {
		"en": "Skin Title"
	},
	"subtitle": {
		"en": "Skin Subtitle"
	},
	"author": {
		"en": "Author"
	},
	"tags": [],
	"description": {
		"en": "Skin Description"
	},
	"meta": {
		"custom": true
	}
}
```

必填字段：`version`、`title`、`subtitle`、`author`、`tags`。

`version` 必须为 `4`。

#### `thumbnail[.png/.srl]`

Skin 缩略图。

#### `data[.json/.srl]`

Skin 数据。

#### `texture[.png/.srl]`

Skin 纹理。

### Backgrounds

每个 background 的资源位于 `backgrounds/{background}`。

#### `item.json`

Background 信息。

```go
type BackgroundItem struct {
	Version     int                       `json:"version"`
	Title       database.LocalizationText `json:"title"`
	Subtitle    database.LocalizationText `json:"subtitle"`
	Author      database.LocalizationText `json:"author"`
	Tags        []database.DatabaseTag    `json:"tags"`
	Description database.LocalizationText `json:"description,omitempty"`
	Meta        json.RawMessage           `json:"meta,omitempty"`
}
```

示例：

```json
{
	"version": 2,
	"title": {
		"en": "Background Title"
	},
	"subtitle": {
		"en": "Background Subtitle"
	},
	"author": {
		"en": "Author"
	},
	"tags": [],
	"description": {
		"en": "Background Description"
	},
	"meta": {
		"custom": true
	}
}
```

必填字段：`version`、`title`、`subtitle`、`author`、`tags`。

`version` 必须为 `2`。

#### `thumbnail[.png/.srl]`

Background 缩略图。

#### `data[.json/.srl]`

Background 数据。

#### `image[.png/.srl]`

Background 图像。

#### `configuration[.json/.srl]`

Background 配置。

### Effects

每个 effect 的资源位于 `effects/{effect}`。

#### `item.json`

Effect 信息。

```go
type EffectItem struct {
	Version     int                       `json:"version"`
	Title       database.LocalizationText `json:"title"`
	Subtitle    database.LocalizationText `json:"subtitle"`
	Author      database.LocalizationText `json:"author"`
	Tags        []database.DatabaseTag    `json:"tags"`
	Description database.LocalizationText `json:"description,omitempty"`
	Meta        json.RawMessage           `json:"meta,omitempty"`
}
```

示例：

```json
{
	"version": 5,
	"title": {
		"en": "Effect Title"
	},
	"subtitle": {
		"en": "Effect Subtitle"
	},
	"author": {
		"en": "Author"
	},
	"tags": [],
	"description": {
		"en": "Effect Description"
	},
	"meta": {
		"custom": true
	}
}
```

必填字段：`version`、`title`、`subtitle`、`author`、`tags`。

`version` 必须为 `5`。

#### `thumbnail[.png/.srl]`

Effect 缩略图。

#### `data[.json/.srl]`

Effect 数据。

#### `audio[.zip/.srl]`

Effect 音频。

### Particles

每个 particle 的资源位于 `particles/{particle}`。

#### `item.json`

Particle 信息。

```go
type ParticleItem struct {
	Version     int                       `json:"version"`
	Title       database.LocalizationText `json:"title"`
	Subtitle    database.LocalizationText `json:"subtitle"`
	Author      database.LocalizationText `json:"author"`
	Tags        []database.DatabaseTag    `json:"tags"`
	Description database.LocalizationText `json:"description,omitempty"`
	Meta        json.RawMessage           `json:"meta,omitempty"`
}
```

示例：

```json
{
	"version": 3,
	"title": {
		"en": "Particle Title"
	},
	"subtitle": {
		"en": "Particle Subtitle"
	},
	"author": {
		"en": "Author"
	},
	"tags": [],
	"description": {
		"en": "Particle Description"
	},
	"meta": {
		"custom": true
	}
}
```

必填字段：`version`、`title`、`subtitle`、`author`、`tags`。

`version` 必须为 `3`。

#### `thumbnail[.png/.srl]`

Particle 缩略图。

#### `data[.json/.srl]`

Particle 数据。

#### `texture[.png/.srl]`

Particle 纹理。

### Engines

每个 engine 的资源位于 `engines/{engine}`。

#### `item.json`

Engine 信息。

```go
type EngineItem struct {
	Version     int                       `json:"version"`
	Title       database.LocalizationText `json:"title"`
	Subtitle    database.LocalizationText `json:"subtitle"`
	Author      database.LocalizationText `json:"author"`
	Tags        []database.DatabaseTag    `json:"tags"`
	Description database.LocalizationText `json:"description,omitempty"`
	Skin        string                    `json:"skin"`
	Background  string                    `json:"background"`
	Effect      string                    `json:"effect"`
	Particle    string                    `json:"particle"`
	Meta        json.RawMessage           `json:"meta,omitempty"`
}
```

示例：

```json
{
	"version": 13,
	"title": {
		"en": "Engine Title"
	},
	"subtitle": {
		"en": "Engine Subtitle"
	},
	"author": {
		"en": "Author"
	},
	"tags": [],
	"description": {
		"en": "Engine Description"
	},
	"skin": "skin-name",
	"background": "background-name",
	"effect": "effect-name",
	"particle": "particle-name",
	"meta": {
		"custom": true
	}
}
```

必填字段：`version`、`title`、`subtitle`、`author`、`tags`、`skin`、`background`、`effect`、`particle`。

`version` 必须为 `13`。`skin`、`background`、`effect`、`particle` 必须能在对应目录中找到。

#### `thumbnail[.png/.srl]`

Engine 缩略图。

#### `playData[.json/.srl]`

Engine play data。

#### `watchData[.json/.srl]`

Engine watch data。

#### `previewData[.json/.srl]`

Engine preview data。

#### `tutorialData[.json/.srl]`

Engine tutorial data。

#### `rom[.bin/.srl]`

可选，engine rom。

#### `configuration[.json/.srl]`

Engine 配置。

### Replays

每个 replay 的资源位于 `replays/{replay}`。

#### `item.json`

Replay 信息。

```go
type ReplayItem struct {
	Version     int                       `json:"version"`
	Title       database.LocalizationText `json:"title"`
	Subtitle    database.LocalizationText `json:"subtitle"`
	Author      database.LocalizationText `json:"author"`
	Tags        []database.DatabaseTag    `json:"tags"`
	Description database.LocalizationText `json:"description,omitempty"`
	Level       string                    `json:"level"`
	Meta        json.RawMessage           `json:"meta,omitempty"`
}
```

示例：

```json
{
	"version": 1,
	"title": {
		"en": "Replay Title"
	},
	"subtitle": {
		"en": "Replay Subtitle"
	},
	"author": {
		"en": "Author"
	},
	"tags": [],
	"description": {
		"en": "Replay Description"
	},
	"level": "level-name",
	"meta": {
		"custom": true
	}
}
```

必填字段：`version`、`title`、`subtitle`、`author`、`tags`、`level`。

`version` 必须为 `1`。`level` 必须能在 `levels/` 中找到。

#### `data[.json/.srl]`

Replay 数据。

#### `configuration[.json/.srl]`

Replay 配置。

## 输出

输出目录默认为 `pack`：

```txt
pack/
  db.json
  repository/
    {hash}
```

`repository/` 包含处理后的资源。`db.json` 包含打包后的 Sonolus 数据库。

打包开始前会清空输出目录；如果打包失败，输出目录会被删除并返回非 0 退出码，避免留下不完整结果。
