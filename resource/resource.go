package resource

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/WindowsSov8forUs/sonolus-core-go/core"
	"github.com/WindowsSov8forUs/sonolus-core-go/crypto"
	"github.com/WindowsSov8forUs/sonolus-pack-go/schema"
)

type Logger interface {
	Info(args ...any)
	Warning(args ...any)
}

type Packer struct {
	Output string
	Logger Logger
}

func (p Packer) Pack(pathBase, ext string, optional bool) (*core.Srl, bool, error) {
	if srl, ok, err := p.packSrl(pathBase); ok || err != nil {
		return srl, ok, err
	}

	pathExt := pathBase + "." + ext
	if data, ok, err := readIfExists(pathBase); ok || err != nil {
		if err != nil {
			return nil, false, fileError(pathBase, err)
		}
		return p.write(data)
	}
	if data, ok, err := readIfExists(pathExt); ok || err != nil {
		if err != nil {
			return nil, false, fileError(pathExt, err)
		}
		switch ext {
		case "json":
			srl, ok, err := p.packJSON(data)
			if err != nil {
				return nil, false, fmt.Errorf("%s: %w", displayPath(pathExt), err)
			}
			return srl, ok, nil
		case "bin":
			return p.packBin(data)
		default:
			return p.write(data)
		}
	}

	if optional {
		if p.Logger != nil {
			p.Logger.Info(fmt.Sprintf("%s[.%s/.srl]: Does not exist, skipped", displayPath(pathBase), ext))
		}
		return nil, false, nil
	}
	if p.Logger != nil {
		p.Logger.Warning(fmt.Sprintf("%s[.%s/.srl]: Does not exist", displayPath(pathBase), ext))
	}
	empty := core.Srl{}
	return &empty, true, nil
}

func (p Packer) packSrl(pathBase string) (*core.Srl, bool, error) {
	path := pathBase + ".srl"
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	srl, err := schema.ParseSrl(path)
	if err != nil {
		return nil, false, err
	}
	return &srl, true, nil
}

func (p Packer) packJSON(data []byte) (*core.Srl, bool, error) {
	packed, err := packJSONResource(data)
	if err != nil {
		return nil, false, err
	}
	return p.write(packed)
}

func (p Packer) packBin(data []byte) (*core.Srl, bool, error) {
	packed, err := packBinResource(data)
	if err != nil {
		return nil, false, err
	}
	return p.write(packed)
}

func (p Packer) write(data []byte) (*core.Srl, bool, error) {
	hash := crypto.Hash(data)
	path := filepath.Join(p.Output, "repository", hash)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, false, fileError(filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return nil, false, fileError(path, err)
	}

	hashValue := core.Value(hash)
	urlValue := core.Value("/sonolus/repository/" + hash)
	return &core.Srl{Hash: &hashValue, URL: &urlValue}, true, nil
}

func readIfExists(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return data, true, nil
	}
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	return nil, false, err
}

func fileError(path string, err error) error {
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		return fmt.Errorf("%s: %s: %s", displayPath(path), pathErr.Op, pathErr.Err)
	}
	return fmt.Errorf("%s: %w", displayPath(path), err)
}

func displayPath(path string) string {
	return filepath.ToSlash(path)
}
