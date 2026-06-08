package resource

import (
	"bytes"
	"compress/gzip"
	"encoding/json"

	"github.com/WindowsSov8forUs/sonolus-core-go/codec"
)

func packJSONResource(raw []byte) ([]byte, error) {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return codec.Compress(value)
}

func packBinResource(raw []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buffer, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write(raw); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
