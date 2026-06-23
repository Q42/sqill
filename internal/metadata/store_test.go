package metadata

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStoreSaveLoad(t *testing.T) {
	dir := t.TempDir()
	s, err := NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	state, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(state.Installed) != 0 {
		t.Fatalf("expected empty state, got %v", state.Installed)
	}

	entry := InstalledEntry{
		Version:     "1.0.0",
		Source:      "https://example.com/skill.git",
		InstalledAt: Now(),
		Description: "test",
	}
	if err := s.Add("demo", entry); err != nil {
		t.Fatalf("add: %v", err)
	}

	got, err := s.Get("demo")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Version != entry.Version || got.Source != entry.Source {
		t.Fatalf("got %+v", got)
	}
	if !s.IsInstalled("demo") {
		t.Fatal("expected installed")
	}

	loaded, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(loaded.Installed) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Installed))
	}
}

func TestFileStoreRemove(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	_ = s.Add("x", InstalledEntry{Version: "1.0.0"})
	if err := s.Remove("x"); err != nil {
		t.Fatal(err)
	}
	if s.IsInstalled("x") {
		t.Fatal("still installed")
	}
	if err := s.Remove("x"); err == nil {
		t.Fatal("expected error removing missing skill")
	}
}

func TestFileStoreAtomicSave(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	_ = s.Add("x", InstalledEntry{Version: "1.0.0"})

	data, err := os.ReadFile(filepath.Join(s.Path()))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Fatal("expected file content")
	}
}

func TestLoadManifest(t *testing.T) {
	dir := t.TempDir()
	m := Manifest{Name: "foo", Version: "1.2.3", Description: "hi"}
	if err := WriteManifest(dir, m); err != nil {
		t.Fatal(err)
	}
	got, err := LoadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != m.Name || got.Version != m.Version || got.Description != m.Description {
		t.Fatalf("got %+v", got)
	}
}

func TestLoadManifestMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := LoadManifest(dir); err == nil {
		t.Fatal("expected error for missing manifest")
	}
}

func TestSortedNames(t *testing.T) {
	state := NewState()
	state.Installed = map[string]InstalledEntry{
		"c": {},
		"a": {},
		"b": {},
	}
	names := SortedNames(state)
	if len(names) != 3 || names[0] != "a" || names[1] != "b" || names[2] != "c" {
		t.Fatalf("got %v", names)
	}
}
