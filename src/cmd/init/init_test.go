package init

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sqill/src/lib/utils"
)

func TestInitCreatesStateFile(t *testing.T) {
	dir := t.TempDir()
	opts := &Options{
		SkillsDir: filepath.Join(dir, ".agents", "skills"),
		Yes:       true,
	}
	if err := runForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	statePath := filepath.Join(opts.SkillsDir, "sqill.json")
	if _, err := os.Stat(statePath); err != nil {
		t.Fatalf("expected state file at %s: %v", statePath, err)
	}
}

func TestInitIdempotent(t *testing.T) {
	dir := t.TempDir()
	opts := &Options{
		SkillsDir: filepath.Join(dir, ".agents", "skills"),
		Yes:       true,
	}
	if err := runForTest(t, opts); err != nil {
		t.Fatal(err)
	}
	if err := runForTest(t, opts); err != nil {
		t.Fatalf("second run should be idempotent: %v", err)
	}
}

func TestInitReportsAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	opts := &Options{
		SkillsDir: filepath.Join(dir, ".agents", "skills"),
		Yes:       true,
	}
	if err := runForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	if err := runCapture(opts, buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "already initialized") {
		t.Fatalf("expected 'already initialized' notice, got:\n%s", buf.String())
	}
}

func TestInitCreatesMissingSymlink(t *testing.T) {
	dir := t.TempDir()
	opts := &Options{
		SkillsDir:  filepath.Join(dir, ".agents", "skills"),
		LinkClaude: true,
	}
	if err := runForTest(t, opts); err != nil {
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
	want, _ := filepath.Rel(filepath.Dir(link), opts.SkillsDir)
	if target != want {
		t.Fatalf("symlink target = %q, want %q", target, want)
	}
}

func TestInitMovesExistingContents(t *testing.T) {
	dir := t.TempDir()
	opts := &Options{
		SkillsDir:  filepath.Join(dir, ".agents", "skills"),
		LinkClaude: true,
	}
	if err := runForTest(t, opts); err != nil {
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

	if err := runForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(opts.SkillsDir, "oldskill", "sqill.json")); err != nil {
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

func TestInitRejectsDuplicates(t *testing.T) {
	dir := t.TempDir()
	opts := &Options{
		SkillsDir:  filepath.Join(dir, ".agents", "skills"),
		LinkClaude: true,
	}
	if err := runForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	existing := filepath.Join(dir, ".claude", "skills")
	if err := os.RemoveAll(existing); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(existing, "dup"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(opts.SkillsDir, "dup"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := runForTest(t, opts)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

func TestInitAlreadySymlink(t *testing.T) {
	dir := t.TempDir()
	opts := &Options{
		SkillsDir:  filepath.Join(dir, ".agents", "skills"),
		LinkClaude: true,
	}
	if err := runForTest(t, opts); err != nil {
		t.Fatal(err)
	}

	if err := runForTest(t, opts); err != nil {
		t.Fatalf("second init should not error on existing symlink: %v", err)
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

	got := utils.DisplayPath(file)
	if got != filepath.Join("foo", "bar", "x.txt") {
		t.Fatalf("expected relative path, got %q", got)
	}
}

func runForTest(t *testing.T, opts *Options) error {
	t.Helper()
	cmd := NewCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--skills-dir", opts.SkillsDir})
	if opts.LinkClaude {
		cmd.SetArgs(append(cmd.Flags().Args(), "--link-claude"))
	}
	if opts.LinkCursor {
		cmd.SetArgs(append(cmd.Flags().Args(), "--link-cursor"))
	}
	if opts.LinkKilo {
		cmd.SetArgs(append(cmd.Flags().Args(), "--link-kilo"))
	}
	if opts.Yes {
		cmd.SetArgs(append(cmd.Flags().Args(), "--yes"))
	}
	if err := cmd.ParseFlags(cmd.Flags().Args()); err != nil {
		return err
	}
	return run(cmd, opts)
}

func runCapture(opts *Options, buf *bytes.Buffer) error {
	cmd := NewCmd()
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--skills-dir", opts.SkillsDir, "--yes"})
	if err := cmd.ParseFlags(cmd.Flags().Args()); err != nil {
		return err
	}
	return run(cmd, opts)
}
