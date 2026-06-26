package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sqill/src/lib/metadata"
)

func TestGuardRequiresStateFile(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, ".agents", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skillsDir, "list"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected guard error")
	}
	if !strings.Contains(err.Error(), "sqill init") {
		t.Fatalf("expected init hint, got %v", err)
	}
}

func TestGuardSkippedForInit(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, ".agents", "skills")

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skillsDir, "init", "--yes"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("init should not require pre-existing state: %v", err)
	}

	if _, err := os.Stat(filepath.Join(skillsDir, "sqill.json")); err != nil {
		t.Fatalf("expected state file: %v", err)
	}
}

func TestInstallRejectsEmptyName(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "install", ""})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestInstallRejectsInvalidChars(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	tests := []string{"../escape", "a/b"}
	for _, name := range tests {
		root.SetArgs([]string{"--skills-dir", skills, "install", name})
		if err := root.Execute(); err == nil {
			t.Fatalf("expected error for %q", name)
		}
	}
}

func TestInstallMissingSkill(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "install", "does-not-exist"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing skill in registry")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestRemoveRejectsEmptyName(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "remove", ""})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRemoveMissingSkill(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "remove", "nonexistent"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing skill")
	}
}

func TestUpdateRejectsEmptyName(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "update", ""})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestUpdateMissingSkill(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "update", "nonexistent"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing skill")
	}
}

func TestListShowsNoSkillsWhenEmpty(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--skills-dir", skills, "list"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "No skills installed") {
		t.Fatalf("expected 'No skills installed', got %q", got)
	}
}

func TestListShowsInstalledSkills(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	store, _ := metadata.NewFileStore(skills)
	_ = store.Add("x", metadata.InstalledEntry{Version: "1.0.0", Source: "local", InstalledAt: metadata.Now()})

	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--skills-dir", skills, "list"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "x") || !strings.Contains(got, "1.0.0") {
		t.Fatalf("expected 'x' and '1.0.0', got %q", got)
	}
}

func TestInfoRequiresNameArg(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "info"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestInfoMissingSkill(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "info", "nonexistent"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing skill")
	}
}

func TestInfoShowsManifest(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(skills, "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, "x", "sqill.json"), []byte(`{"name":"x","version":"1.2.3","description":"desc"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	store, _ := metadata.NewFileStore(skills)
	_ = store.Add("x", metadata.InstalledEntry{Version: "1.2.3", Source: "file:///src", InstalledAt: metadata.Now()})

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "info", "x"})
	if err := root.Execute(); err != nil {
		t.Fatalf("info failed: %v", err)
	}
}

func TestInitWritesGitignore(t *testing.T) {
	skills := t.TempDir()

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "init", "--yes"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(skills, metadata.GitignoreFileName))
	if err != nil {
		t.Fatalf("expected .gitignore: %v", err)
	}
	if !strings.Contains(string(data), "Managed by sqill") {
		t.Fatalf("missing header, got %q", string(data))
	}
}

func TestTrackUpdatesGitignore(t *testing.T) {
	skills := t.TempDir()
	seedSkillDir(t, skills, "x")
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	store, _ := metadata.NewFileStore(skills)
	_ = store.Add("x", metadata.InstalledEntry{Version: "1.0.0", Source: "local", InstalledAt: metadata.Now()})

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "track", "x"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("track: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(skills, metadata.GitignoreFileName))
	if err != nil {
		t.Fatalf("expected .gitignore: %v", err)
	}
	if strings.Contains(string(data), "x/") {
		t.Fatalf("x should not be ignored after track, got %q", string(data))
	}
}

func TestUntrackUpdatesGitignore(t *testing.T) {
	skills := t.TempDir()
	seedSkillDir(t, skills, "x")
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	store, _ := metadata.NewFileStore(skills)
	_ = store.Add("x", metadata.InstalledEntry{Version: "1.0.0", Source: "local", InstalledAt: metadata.Now()})
	_ = store.Track("x")
	_ = metadata.SyncGitignore(skills)

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "untrack", "x"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("untrack: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(skills, metadata.GitignoreFileName))
	if err != nil {
		t.Fatalf("expected .gitignore: %v", err)
	}
	if !strings.Contains(string(data), "x/\n") {
		t.Fatalf("x should be ignored after untrack, got %q", string(data))
	}
}

func TestTrackRejectsUnknownSkill(t *testing.T) {
	skills := t.TempDir()
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skills, "track", "missing"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error tracking non-installed skill")
	}
}

func seedSkillDir(t *testing.T, skillsDir, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(skillsDir, name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, name, "sqill.json"), []byte(`{"name":"`+name+`","version":"1.0.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInstallNoArgsEmptyState(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(skills, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--skills-dir", skills, "install"})
	if err := root.Execute(); err != nil {
		t.Fatalf("install: %v", err)
	}
	if !strings.Contains(buf.String(), "No skills to install") {
		t.Fatalf("expected message, got %q", buf.String())
	}
}

func TestVersionFlag(t *testing.T) {
	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("--version: %v", err)
	}
	if !strings.Contains(buf.String(), "sqill version") {
		t.Fatalf("expected version output, got %q", buf.String())
	}
}

func TestUpgradeDoesNotRequireState(t *testing.T) {
	root := NewRoot()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--skills-dir", "/no/such/dir", "upgrade", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("upgrade --help should not require state file: %v", err)
	}
}