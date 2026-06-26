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

func TestInstallSubdirResolution(t *testing.T) {
	cases := []struct {
		name    string
		subdir  string
		version string
	}{
		{"named", "x", "0.1.0"},
		{"skill", "skill", "0.2.0"},
		{"sqill", "sqill", "0.3.0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := t.TempDir()
			inner := filepath.Join(src, tc.subdir)
			if err := os.MkdirAll(inner, 0o755); err != nil {
				t.Fatal(err)
			}
			manifest := `{"name":"x","version":"` + tc.version + `","description":"d"}`
			if err := os.WriteFile(filepath.Join(inner, "sqill.json"), []byte(manifest), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(inner, "SKILL.md"), []byte("hello"), 0o644); err != nil {
				t.Fatal(err)
			}

			skills := t.TempDir()
			store, _ := metadata.NewFileStore(skills)
			reg := &fakeReg{source: "file://" + src}
			inst := New(reg, store, skills)

			if err := inst.Install("x", false); err != nil {
				t.Fatalf("install: %v", err)
			}

			if _, err := os.Stat(filepath.Join(skills, "x", "sqill.json")); err != nil {
				t.Fatalf("sqill.json not flattened to root: %v", err)
			}
			if _, err := os.Stat(filepath.Join(skills, "x", "SKILL.md")); err != nil {
				t.Fatalf("SKILL.md not flattened to root: %v", err)
			}
			if _, err := os.Stat(filepath.Join(skills, "x", tc.subdir)); !os.IsNotExist(err) {
				t.Fatalf("subdir %q should have been removed, got %v", tc.subdir, err)
			}
			if !store.IsInstalled("x") {
				t.Fatal("expected installed")
			}
			got, _ := store.Get("x")
			if got.Version != tc.version {
				t.Fatalf("expected %s, got %s", tc.version, got.Version)
			}
		})
	}
}

func TestUpdateSubdirResolution(t *testing.T) {
	src := t.TempDir()
	inner := filepath.Join(src, "x")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inner, "sqill.json"), []byte(`{"name":"x","version":"1.0.0","description":"d"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inner, "SKILL.md"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	skills := t.TempDir()
	store, _ := metadata.NewFileStore(skills)
	reg := &fakeReg{source: "file://" + src}
	inst := New(reg, store, skills)

	if err := inst.Install("x", false); err != nil {
		t.Fatalf("install: %v", err)
	}

	if err := os.WriteFile(filepath.Join(inner, "sqill.json"), []byte(`{"name":"x","version":"2.0.0","description":"d"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := inst.Update("x"); err != nil {
		t.Fatalf("update: %v", err)
	}

	if _, err := os.Stat(filepath.Join(skills, "x", "sqill.json")); err != nil {
		t.Fatalf("sqill.json not flattened: %v", err)
	}
	got, _ := store.Get("x")
	if got.Version != "2.0.0" {
		t.Fatalf("expected 2.0.0, got %s", got.Version)
	}
}

func TestInstallNoManifestAnywhere(t *testing.T) {
	src := t.TempDir()
	for _, sub := range []string{"x", "skill", "sqill"} {
		if err := os.MkdirAll(filepath.Join(src, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	skills := t.TempDir()
	store, _ := metadata.NewFileStore(skills)
	reg := &fakeReg{source: "file://" + src}
	inst := New(reg, store, skills)

	err := inst.Install("x", false)
	if err == nil {
		t.Fatal("expected error when no sqill.json anywhere")
	}
	if _, statErr := os.Stat(filepath.Join(skills, "x")); !os.IsNotExist(statErr) {
		t.Fatalf("target dir should have been cleaned up, got %v", statErr)
	}
	if store.IsInstalled("x") {
		t.Fatal("should not be marked installed")
	}
}

func TestPickManifestDir(t *testing.T) {
	root := t.TempDir()
	mustWrite := func(p, body string) {
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("root wins over subdirs", func(t *testing.T) {
		mustWrite(filepath.Join(root, "sqill.json"), `{"name":"x"}`)
		if err := os.MkdirAll(filepath.Join(root, "x"), 0o755); err != nil {
			t.Fatal(err)
		}
		mustWrite(filepath.Join(root, "x", "sqill.json"), `{"name":"x"}`)
		got, err := pickManifestDir(root, "x")
		if err != nil {
			t.Fatal(err)
		}
		if got != root {
			t.Fatalf("expected root, got %s", got)
		}
	})

	t.Run("named subdir before skill before sqill", func(t *testing.T) {
		dir := t.TempDir()
		for _, sub := range []string{"x", "skill", "sqill"} {
			if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
				t.Fatal(err)
			}
			mustWrite(filepath.Join(dir, sub, "sqill.json"), `{"name":"x"}`)
		}
		got, err := pickManifestDir(dir, "x")
		if err != nil {
			t.Fatal(err)
		}
		if got != filepath.Join(dir, "x") {
			t.Fatalf("expected named subdir first, got %s", got)
		}
	})

	t.Run("skill before sqill", func(t *testing.T) {
		dir := t.TempDir()
		for _, sub := range []string{"skill", "sqill"} {
			if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
				t.Fatal(err)
			}
			mustWrite(filepath.Join(dir, sub, "sqill.json"), `{"name":"x"}`)
		}
		got, err := pickManifestDir(dir, "other")
		if err != nil {
			t.Fatal(err)
		}
		if got != filepath.Join(dir, "skill") {
			t.Fatalf("expected skill before sqill, got %s", got)
		}
	})

	t.Run("missing", func(t *testing.T) {
		dir := t.TempDir()
		if _, err := pickManifestDir(dir, "x"); err == nil {
			t.Fatal("expected error")
		}
	})
}