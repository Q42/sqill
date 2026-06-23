package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupCreatesStateFile(t *testing.T) {
	dir := t.TempDir()
	opts := &setupOptions{
		skillsDir: filepath.Join(dir, ".agents", "skills"),
		yes:       true,
	}
	if err := runSetupForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	statePath := filepath.Join(opts.skillsDir, "sqill.json")
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("expected state file at %s: %v", statePath, err)
	}
}

func TestSetupIdempotent(t *testing.T) {
	dir := t.TempDir()
	opts := &setupOptions{
		skillsDir: filepath.Join(dir, ".agents", "skills"),
		yes:       true,
	}
	if err := runSetupForTest(t, opts); err != nil {
		t.Fatal(err)
	}
	if err := runSetupForTest(t, opts); err != nil {
		t.Fatalf("second run should be idempotent: %v", err)
	}
}

func TestSetupCreatesMissingSymlink(t *testing.T) {
	dir := t.TempDir()
	opts := &setupOptions{
		skillsDir:  filepath.Join(dir, ".agents", "skills"),
		linkClaude: true,
	}
	if err := runSetupForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, ".claude", "skills")
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("expected symlink: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected symlink, got regular entry")
	}

	target, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Rel(filepath.Dir(link), opts.skillsDir)
	if target != want {
		t.Fatalf("symlink target = %q, want %q", target, want)
	}
}

func TestSetupMovesExistingContents(t *testing.T) {
	dir := t.TempDir()
	opts := &setupOptions{
		skillsDir:  filepath.Join(dir, ".agents", "skills"),
		linkClaude: true,
	}
	if err := runSetupForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	link := filepath.Join(dir, ".claude", "skills")
	if err := os.RemoveAll(link); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(dir, ".claude", "skills")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	oldSkill := filepath.Join(existing, "oldskill")
	if err := os.MkdirAll(oldSkill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(oldSkill, "sqill.json"), []byte(`{"name":"oldskill","version":"1.0.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runSetupForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(opts.skillsDir, "oldskill", "sqill.json")); err != nil {
		t.Fatalf("expected moved skill: %v", err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("expected symlink: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("expected symlink after move")
	}
}

func TestSetupRejectsDuplicates(t *testing.T) {
	dir := t.TempDir()
	opts := &setupOptions{
		skillsDir:  filepath.Join(dir, ".agents", "skills"),
		linkClaude: true,
	}
	if err := runSetupForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	existing := filepath.Join(dir, ".claude", "skills")
	if err := os.RemoveAll(existing); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(existing, "dup"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(opts.skillsDir, "dup"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := runSetupForTest(t, opts)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestSetupAlreadySymlink(t *testing.T) {
	dir := t.TempDir()
	opts := &setupOptions{
		skillsDir:  filepath.Join(dir, ".agents", "skills"),
		linkClaude: true,
	}
	if err := runSetupForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	if err := runSetupForTest(t, opts); err != nil {
		t.Fatalf("second setup should not error on existing symlink: %v", err)
	}
}

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
	if !strings.Contains(err.Error(), "sqill setup") {
		t.Fatalf("expected setup hint, got %v", err)
	}
}

func TestGuardSkippedForSetup(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, ".agents", "skills")

	root := NewRoot()
	root.SetArgs([]string{"--skills-dir", skillsDir, "setup", "--yes"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err != nil {
		t.Fatalf("setup should not require pre-existing state: %v", err)
	}

	if _, err := os.Stat(filepath.Join(skillsDir, "sqill.json")); err != nil {
		t.Fatalf("expected state file: %v", err)
	}
}

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

	got := displayPath(file)
	if got != filepath.Join("foo", "bar", "x.txt") {
		t.Fatalf("expected relative path, got %q", got)
	}
}

func runSetupForTest(t *testing.T, opts *setupOptions) error {
	t.Helper()
	cmd := newSetupCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--skills-dir", opts.skillsDir})
	if opts.linkClaude {
		cmd.SetArgs(append(cmd.Flags().Args(), "--link-claude"))
	}
	if opts.linkCursor {
		cmd.SetArgs(append(cmd.Flags().Args(), "--link-cursor"))
	}
	if opts.linkKilo {
		cmd.SetArgs(append(cmd.Flags().Args(), "--link-kilo"))
	}
	if opts.yes {
		cmd.SetArgs(append(cmd.Flags().Args(), "--yes"))
	}
	if err := cmd.ParseFlags(cmd.Flags().Args()); err != nil {
		return err
	}
	return runSetup(cmd, opts)
}
