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

func TestSearchShowsNoMatches(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(filepath.Join(skills, ".agents", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, ".agents", "skills", metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--skills-dir", filepath.Join(skills, ".agents", "skills"), "search", "nosuchskill"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "No matches") {
		t.Fatalf("expected 'No matches', got %q", got)
	}
}

func TestSearchFindsMatch(t *testing.T) {
	skills := t.TempDir()
	if err := os.MkdirAll(filepath.Join(skills, ".agents", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skills, ".agents", "skills", metadata.StateFileName), []byte(`{"installed":{},"registries":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	root := NewRoot()
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--skills-dir", filepath.Join(skills, ".agents", "skills"), "search", "sRegressor"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "sRegressor") {
		t.Fatalf("expected 'sRegressor', got %q", got)
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