package installer

import (
	"os"
	"path/filepath"
	"testing"

	"sqill/src/lib/metadata"
	"sqill/src/lib/registry"
	"sqill/src/lib/utils"
)

func TestValidateName(t *testing.T) {
	good := []string{"abc", "x1", "with-dash", "with_under"}
	for _, n := range good {
		if err := utils.ValidateName(n); err != nil {
			t.Errorf("expected ok for %q, got %v", n, err)
		}
	}
	bad := []string{"", ".dot", "..", "../escape", "a/b", `a\b`, "a..b"}
	for _, n := range bad {
		if err := utils.ValidateName(n); err == nil {
			t.Errorf("expected error for %q", n)
		}
	}
}

type fakeReg struct {
	source string
	desc   string
}

func (f *fakeReg) Search(q string) []registry.SkillEntry {
	return []registry.SkillEntry{{Name: "x", Source: f.source, Description: f.desc}}
}

func (f *fakeReg) Resolve(n string) (registry.SkillEntry, error) {
	return registry.SkillEntry{Name: n, Source: f.source, Description: f.desc}, nil
}

func (f *fakeReg) All() []registry.SkillEntry {
	return []registry.SkillEntry{{Name: "x", Source: f.source, Description: f.desc}}
}

func TestInstallAndRemoveLocal(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "sqill.json"), []byte(`{"name":"x","version":"1.2.3","description":"hi"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	skills := t.TempDir()
	store, err := metadata.NewFileStore(skills)
	if err != nil {
		t.Fatal(err)
	}

	reg := &fakeReg{source: "file://" + src, desc: "hi"}
	inst := New(reg, store, skills)

	if err := inst.Install("x", false); err != nil {
		t.Fatalf("install: %v", err)
	}

	if !store.IsInstalled("x") {
		t.Fatal("expected installed")
	}

	if _, err := os.Stat(filepath.Join(skills, "x", "sqill.json")); err != nil {
		t.Fatalf("expected sqill.json on disk: %v", err)
	}

	if err := inst.Install("x", false); err == nil {
		t.Fatal("expected error on duplicate install")
	}

	if err := inst.Install("x", true); err != nil {
		t.Fatalf("force reinstall: %v", err)
	}

	if err := inst.Remove("x"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if store.IsInstalled("x") {
		t.Fatal("still installed after remove")
	}
	if _, err := os.Stat(filepath.Join(skills, "x")); !os.IsNotExist(err) {
		t.Fatalf("expected dir gone, got %v", err)
	}
}

func TestInstallManifestMismatch(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "sqill.json"), []byte(`{"name":"y","version":"1.0.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	skills := t.TempDir()
	store, _ := metadata.NewFileStore(skills)
	reg := &fakeReg{source: "file://" + src}
	inst := New(reg, store, skills)

	err := inst.Install("x", false)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
	if _, statErr := os.Stat(filepath.Join(skills, "x")); !os.IsNotExist(statErr) {
		t.Fatal("target dir should have been cleaned up")
	}
}

func TestUpdate(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "sqill.json"), []byte(`{"name":"x","version":"1.0.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	skills := t.TempDir()
	store, _ := metadata.NewFileStore(skills)
	reg := &fakeReg{source: "file://" + src}
	inst := New(reg, store, skills)

	if err := inst.Install("x", false); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(src, "sqill.json"), []byte(`{"name":"x","version":"2.0.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := inst.Update("x"); err != nil {
		t.Fatal(err)
	}

	got, _ := store.Get("x")
	if got.Version != "2.0.0" {
		t.Fatalf("expected 2.0.0, got %s", got.Version)
	}
}