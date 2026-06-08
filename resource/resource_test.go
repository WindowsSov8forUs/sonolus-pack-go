package resource_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/WindowsSov8forUs/sonolus-core-go/crypto"
	"github.com/WindowsSov8forUs/sonolus-pack-go/resource"
)

func TestPackUsesSrlPriority(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "pack")
	base := filepath.Join(dir, "resource")
	write(t, base+".srl", `{"hash":"external","url":"https://example.com"}`)
	write(t, base+".png", "png")

	srl, ok, err := resource.Packer{Output: output}.Pack(base, "png", false)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || srl == nil || srl.Hash == nil {
		t.Fatalf("expected SRL result: %#v", srl)
	}
	hash, valid := srl.Hash.Value()
	if !valid || hash != "external" {
		t.Fatalf("hash = %q, %v; want external, true", hash, valid)
	}
	if _, err := os.Stat(filepath.Join(output, "repository")); !os.IsNotExist(err) {
		t.Fatalf("repository should not be written for .srl priority, stat err: %v", err)
	}
}

func TestPackUsesExtensionlessBeforeExt(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "pack")
	base := filepath.Join(dir, "resource")
	write(t, base, "raw")
	write(t, base+".png", "png")

	srl, ok, err := resource.Packer{Output: output}.Pack(base, "png", false)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || srl == nil || srl.Hash == nil {
		t.Fatalf("expected SRL result: %#v", srl)
	}
	want := crypto.Hash([]byte("raw"))
	hash, valid := srl.Hash.Value()
	if !valid || hash != want {
		t.Fatalf("hash = %q, %v; want %s, true", hash, valid, want)
	}
	if got, err := os.ReadFile(filepath.Join(output, "repository", want)); err != nil || string(got) != "raw" {
		t.Fatalf("repository content = %q, %v; want raw, nil", got, err)
	}
}

func TestPackOptionalAndRequiredMissing(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "pack")
	base := filepath.Join(dir, "missing")
	packer := resource.Packer{Output: output}

	if srl, ok, err := packer.Pack(base, "png", true); err != nil || ok || srl != nil {
		t.Fatalf("optional missing = %#v, %v, %v; want nil, false, nil", srl, ok, err)
	}
	srl, ok, err := packer.Pack(base, "png", false)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || srl == nil || srl.Hash != nil || srl.URL != nil {
		t.Fatalf("required missing = %#v, %v; want empty SRL, true", srl, ok)
	}
}

func TestPackJSONResourceCompressesJSON(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "pack")
	base := filepath.Join(dir, "data")
	write(t, base+".json", `{"b":2,"a":1}`)

	srl, ok, err := resource.Packer{Output: output}.Pack(base, "json", false)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || srl == nil || srl.Hash == nil {
		t.Fatalf("expected SRL result: %#v", srl)
	}
	hash, _ := srl.Hash.Value()
	data, err := os.ReadFile(filepath.Join(output, "repository", hash))
	if err != nil {
		t.Fatal(err)
	}
	if got := gunzip(t, data); got == "" {
		t.Fatal("expected non-empty decompressed JSON")
	}
}

func TestPackBinResourceGzipsBytes(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "pack")
	base := filepath.Join(dir, "rom")
	write(t, base+".bin", "rom")

	srl, ok, err := resource.Packer{Output: output}.Pack(base, "bin", false)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || srl == nil || srl.Hash == nil {
		t.Fatalf("expected SRL result: %#v", srl)
	}
	hash, _ := srl.Hash.Value()
	data, err := os.ReadFile(filepath.Join(output, "repository", hash))
	if err != nil {
		t.Fatal(err)
	}
	if got := gunzip(t, data); got != "rom" {
		t.Fatalf("decompressed bin = %q; want rom", got)
	}
}

func gunzip(t *testing.T, data []byte) string {
	t.Helper()

	reader, err := gzip.NewReader(bytesReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

func bytesReader(data []byte) *bytes.Reader {
	return bytes.NewReader(data)
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
