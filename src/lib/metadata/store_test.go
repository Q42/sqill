package metadata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestFileStoreTrackUntrack(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	_ = s.Add("foo", InstalledEntry{Version: "1.0.0"})
	_ = s.Add("bar", InstalledEntry{Version: "1.0.0"})

	if err := s.Track("foo"); err != nil {
		t.Fatal(err)
	}
	if !s.IsTracked("foo") {
		t.Fatal("foo should be tracked")
	}
	if s.IsTracked("bar") {
		t.Fatal("bar should not be tracked")
	}

	if err := s.Track("foo"); err != nil {
		t.Fatalf("re-tracking should be idempotent, got %v", err)
	}

	state, _ := s.Load()
	if !state.Installed["foo"].Tracked {
		t.Fatalf("expected foo.Tracked=true, got %+v", state.Installed["foo"])
	}
	if state.Installed["bar"].Tracked {
		t.Fatalf("expected bar.Tracked=false, got %+v", state.Installed["bar"])
	}

	if err := s.Untrack("foo"); err != nil {
		t.Fatal(err)
	}
	if s.IsTracked("foo") {
		t.Fatal("foo should be untracked")
	}
	if err := s.Untrack("foo"); err != nil {
		t.Fatalf("untracking untracked skill should be idempotent, got %v", err)
	}
}

func TestFileStoreSaveAlwaysEmitsTracked(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	_ = s.Add("foo", InstalledEntry{Version: "1.0.0"})
	_ = s.Track("bar")

	data, err := os.ReadFile(filepath.Join(dir, StateFileName))
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Installed map[string]InstalledEntry `json:"installed"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	raw := string(data)
	for _, name := range []string{"foo", "bar"} {
		if !strings.Contains(raw, `"`+name+`"`) {
			continue
		}
		if !strings.Contains(raw, `"tracked":`) {
			t.Fatalf("expected `tracked` key present for %s, got %s", name, raw)
		}
	}
}

func TestFileStoreUntrackRejectsMissing(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	if err := s.Untrack("missing"); err == nil {
		t.Fatal("expected error untracking non-installed skill")
	}
}

func TestFileStoreTrackRejectsMissing(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	if err := s.Track("missing"); err == nil {
		t.Fatal("expected error tracking non-installed skill")
	}
}

func TestFileStoreTrackRejectsInvalidName(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	if err := s.Track("../escape"); err == nil {
		t.Fatal("expected error for invalid name")
	}
}

func TestFileStoreRemoveClearsTracked(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	_ = s.Add("foo", InstalledEntry{Version: "1.0.0"})
	if err := s.Track("foo"); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove("foo"); err != nil {
		t.Fatal(err)
	}
	if s.IsTracked("foo") {
		t.Fatal("foo should not be tracked after remove")
	}
}

func TestSyncGitignoreDefaultIgnoresAll(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	_ = s.Add("foo", InstalledEntry{Version: "1.0.0"})
	_ = s.Add("bar", InstalledEntry{Version: "1.0.0"})

	if err := SyncGitignore(dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, GitignoreFileName))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "Managed by sqill") {
		t.Fatalf("missing header, got %q", content)
	}
	if !strings.Contains(content, "foo/\n") || !strings.Contains(content, "bar/\n") {
		t.Fatalf("expected foo/ and bar/ ignored, got %q", content)
	}
}

func TestSyncGitignoreExcludesTracked(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	_ = s.Add("foo", InstalledEntry{Version: "1.0.0"})
	_ = s.Add("bar", InstalledEntry{Version: "1.0.0"})
	_ = s.Track("foo")

	if err := SyncGitignore(dir); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, GitignoreFileName))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if strings.Contains(content, "foo/") {
		t.Fatalf("foo should not be ignored, got %q", content)
	}
	if !strings.Contains(content, "bar/\n") {
		t.Fatalf("bar should be ignored, got %q", content)
	}
}

func TestSyncGitignoreEmptyInstalls(t *testing.T) {
	dir := t.TempDir()
	_, _ = NewFileStore(dir)
	if err := SyncGitignore(dir); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, GitignoreFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Managed by sqill") {
		t.Fatalf("missing header, got %q", string(data))
	}
}

func TestSyncGitignoreSkipsUnknownTracked(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewFileStore(dir)
	state, _ := s.Load()
	state.Installed["ghost"] = InstalledEntry{Version: "0.0.0", Tracked: true}
	if err := s.Save(state); err != nil {
		t.Fatal(err)
	}
	if err := SyncGitignore(dir); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, GitignoreFileName))
	if strings.Contains(string(data), "ghost") {
		t.Fatalf("tracked-but-uninstalled skill should not appear, got %q", string(data))
	}
}

func TestFileStoreLoadMigratesLegacyTracked(t *testing.T) {
	dir := t.TempDir()
	legacy := `{"installed":{"a":{"version":"1.0.0","source":"x","installed_at":"now"},"b":{"version":"1.0.0","source":"y","installed_at":"now"}},"registries":[],"tracked":["a"]}`
	if err := os.WriteFile(filepath.Join(dir, StateFileName), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	s, _ := NewFileStore(dir)
	state, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !state.Installed["a"].Tracked {
		t.Fatal("a should be tracked after migration")
	}
	if state.Installed["b"].Tracked {
		t.Fatal("b should not be tracked after migration")
	}
}
