package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"sqill/src/lib/installer"
	"sqill/src/lib/metadata"
	"sqill/src/lib/registry"
)

func TestNewCreatesStore(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, ".agents", "skills")
	rt, err := New(skillsDir)
	if err != nil {
		t.Fatal(err)
	}
	if rt.SkillsDir == "" {
		t.Fatal("expected non-empty skills dir")
	}
	if rt.Store == nil {
		t.Fatal("expected non-nil store")
	}
	if rt.Inst == nil {
		t.Fatal("expected non-nil installer")
	}
	if rt.Reg == nil {
		t.Fatal("expected non-nil registry")
	}
	if _, err := os.Stat(skillsDir); err != nil {
		t.Fatalf("expected skills dir at %s: %v", skillsDir, err)
	}
}

func TestNewCreatesRegistry(t *testing.T) {
	dir := t.TempDir()
	rt, err := New(filepath.Join(dir, ".agents", "skills"))
	if err != nil {
		t.Fatal(err)
	}
	entries := rt.Reg.All()
	if len(entries) == 0 {
		t.Fatal("expected at least one entry in registry")
	}
}

func TestNewInstallerForTest(t *testing.T) {
	dir := t.TempDir()
	store, _ := metadata.NewFileStore(dir)
	reg := registry.NewHardcoded()
	inst := installer.New(reg, store, dir)
	if inst == nil {
		t.Fatal("expected non-nil installer")
	}
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
}