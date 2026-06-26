package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDisplayPathRelative(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "foo", "bar")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(subdir, "x.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWd)

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	got := DisplayPath(file)
	if got != filepath.Join("foo", "bar", "x.txt") {
		t.Fatalf("expected relative path, got %q", got)
	}
}

func TestSafeJoin(t *testing.T) {
	root := t.TempDir()
	good, err := SafeJoin(root, "a/b.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(good) {
		t.Fatalf("expected absolute, got %q", good)
	}
	if _, err := SafeJoin(root, "../escape"); err == nil {
		t.Fatal("expected error for path traversal")
	}
}

func TestValidateName(t *testing.T) {
	good := []string{"abc", "x1", "with-dash", "with_under"}
	for _, n := range good {
		if err := ValidateName(n); err != nil {
			t.Errorf("expected ok for %q, got %v", n, err)
		}
	}
	bad := []string{"", ".dot", "..", "../escape", "a/b", `a\b`, "a..b"}
	for _, n := range bad {
		if err := ValidateName(n); err == nil {
			t.Errorf("expected error for %q", n)
		}
	}
}

func TestSubdirNames(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "b"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "c.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	names, err := SubdirNames(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 dirs, got %d", len(names))
	}
}

func TestFindDuplicates(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	if err := os.MkdirAll(filepath.Join(a, "dup"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(b, "dup"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(a, "unique"), 0o755); err != nil {
		t.Fatal(err)
	}
	dupes, err := FindDuplicates(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(dupes) != 1 || dupes[0] != "dup" {
		t.Fatalf("expected [dup], got %v", dupes)
	}
}

func TestStripGitDirs(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git", "objects"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "HEAD"), []byte("ref"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "sub.txt"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := StripGitDirs(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); !os.IsNotExist(err) {
		t.Fatalf("expected .git removed at root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub", ".git")); !os.IsNotExist(err) {
		t.Fatalf("expected .git removed in subdir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "keep.txt")); err != nil {
		t.Fatalf("expected keep.txt preserved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub", "sub.txt")); err != nil {
		t.Fatalf("expected sub/sub.txt preserved: %v", err)
	}
}

func TestMoveContents(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := MoveContents(src, dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dst, "a.txt")); err != nil {
		t.Fatalf("expected a.txt moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "subdir")); err != nil {
		t.Fatalf("expected subdir moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(src, "a.txt")); !os.IsNotExist(err) {
		t.Fatal("expected src empty")
	}
}