package source

import (
	"os"
	"path/filepath"
	"testing"

	"sqill/src/lib/utils"
)

func TestDetect(t *testing.T) {
	cases := []struct {
		src     string
		want    Type
		wantErr bool
	}{
		{"git@github.com:org/x.git", TypeGit, false},
		{"https://github.com/org/x.git", TypeGit, false},
		{"file:///opt/skills/x", TypeLocal, false},
		{"https://example.com/x.tar.gz", TypeArchive, false},
		{"https://example.com/x.tgz", TypeArchive, false},
		{"ftp://example.com/x", "", true},
		{"", "", true},
	}
	for _, c := range cases {
		got, err := Detect(c.src)
		if c.wantErr {
			if err == nil {
				t.Errorf("Detect(%q): expected error", c.src)
			}
			continue
		}
		if err != nil {
			t.Errorf("Detect(%q): %v", c.src, err)
			continue
		}
		if got != c.want {
			t.Errorf("Detect(%q) = %q, want %q", c.src, got, c.want)
		}
	}
}

func TestLocalFetch(t *testing.T) {
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "hello.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "x.txt"), []byte("yo"), 0o644); err != nil {
		t.Fatal(err)
	}

	parent := t.TempDir()
	dest := filepath.Join(parent, "out")

	l := NewLocal()
	if err := l.Fetch("file://"+src, dest); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dest, "hello.txt")); err != nil {
		t.Fatalf("missing hello.txt: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "sub", "x.txt")); err != nil {
		t.Fatalf("missing sub/x.txt: %v", err)
	}
}

func TestLocalFetchSkipsGit(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, ".git", "HEAD"), []byte("ref"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "hello.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	parent := t.TempDir()
	dest := filepath.Join(parent, "out")
	l := NewLocal()
	if err := l.Fetch("file://"+src, dest); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dest, ".git")); !os.IsNotExist(err) {
		t.Fatalf("expected .git to be skipped: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "hello.txt")); err != nil {
		t.Fatalf("missing hello.txt: %v", err)
	}
}

func TestLocalFetchRejectsFile(t *testing.T) {
	src := t.TempDir()
	f := filepath.Join(src, "a.txt")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	parent := t.TempDir()
	dest := filepath.Join(parent, "out")
	l := NewLocal()
	if err := l.Fetch("file://"+f, dest); err == nil {
		t.Fatal("expected error for file source")
	}
}

func TestExtractTarGz(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "test.tar.gz")
	dest := filepath.Join(dir, "out")

	if err := writeSampleTarGz(archive); err != nil {
		t.Fatal(err)
	}
	if err := extractTarGz(archive, dest); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dest, "hello.txt")); err != nil {
		t.Fatalf("missing hello.txt: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "nested", "x.txt")); err != nil {
		t.Fatalf("missing nested/x.txt: %v", err)
	}
}

func TestSafeJoin(t *testing.T) {
	root := t.TempDir()
	good, err := utils.SafeJoin(root, "a/b.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(good) {
		t.Fatalf("expected absolute, got %q", good)
	}
	if _, err := utils.SafeJoin(root, "../escape"); err == nil {
		t.Fatal("expected error for path traversal")
	}
}
