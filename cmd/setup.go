package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sqill/internal/metadata"
)

type linkTarget struct {
	name  string
	path  string
	label string
}

func defaultLinkTargets(baseDir string) []linkTarget {
	return []linkTarget{
		{name: "claude", path: filepath.Join(baseDir, ".claude", "skills"), label: "Claude (.claude/skills)"},
		{name: "cursor", path: filepath.Join(baseDir, ".cursor", "skills"), label: "Cursor (.cursor/skills)"},
		{name: "kilo", path: filepath.Join(baseDir, ".kilo", "skills"), label: "Kilo (.kilo/skills)"},
	}
}

type setupOptions struct {
	skillsDir  string
	linkClaude bool
	linkCursor bool
	linkKilo   bool
	yes        bool
}

func newSetupCmd() *cobra.Command {
	opts := &setupOptions{}

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Initialize .agents/skills and optionally create tool symlinks",
		Long:  "Creates .agents/skills/sqill.json if missing, then optionally symlinks .claude/skills, .cursor/skills, and .kilo/skills into it.",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cmd.Flags().GetString("skills-dir")
			if err != nil {
				return err
			}
			opts.skillsDir = dir
			return runSetup(cmd, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.linkClaude, "link-claude", false, "Symlink .claude/skills into .agents/skills")
	cmd.Flags().BoolVar(&opts.linkCursor, "link-cursor", false, "Symlink .cursor/skills into .agents/skills")
	cmd.Flags().BoolVar(&opts.linkKilo, "link-kilo", false, "Symlink .kilo/skills into .agents/skills")
	cmd.Flags().BoolVar(&opts.yes, "yes", false, "Skip prompts (no symlinks by default)")

	return cmd
}

func runSetup(cmd *cobra.Command, opts *setupOptions) error {
	out := cmd.OutOrStderr()

	skillsDir, err := filepath.Abs(opts.skillsDir)
	if err != nil {
		return fmt.Errorf("resolve skills dir: %w", err)
	}
	skillsDirRel, _ := filepath.Rel(cwdEval(), skillsDir)

	statePath := filepath.Join(skillsDir, metadata.StateFileName)
	stateExists := false
	if _, err := os.Stat(statePath); err == nil {
		stateExists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check state file: %w", err)
	}

	baseDir := filepath.Dir(filepath.Dir(skillsDir))
	targets := defaultLinkTargets(baseDir)

	flagsProvided := opts.linkClaude || opts.linkCursor || opts.linkKilo
	interactive := !flagsProvided && !opts.yes

	if interactive {
		if err := promptSetup(cmd, skillsDirRel, targets, opts); err != nil {
			return err
		}
	} else if opts.yes && !flagsProvided {
		opts.linkClaude = false
		opts.linkCursor = false
		opts.linkKilo = false
	}

	if !stateExists {
		if err := os.MkdirAll(skillsDir, 0o755); err != nil {
			return fmt.Errorf("create skills dir: %w", err)
		}
		store, err := metadata.NewFileStore(skillsDir)
		if err != nil {
			return err
		}
		if err := store.Save(metadata.NewState()); err != nil {
			return fmt.Errorf("write state: %w", err)
		}
		fmt.Fprintf(out, "✓ Created %s\n", displayPath(statePath))
	} else {
		fmt.Fprintf(out, "• %s already exists\n", displayPath(statePath))
	}

	planned := plannedLinks(opts, targets)
	if len(planned) == 0 {
		fmt.Fprintln(out, "• No symlinks requested")
		return nil
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Setting up symlinks:")
	for _, t := range planned {
		fmt.Fprintf(out, "  → %s\n", displayPath(t.path))
	}
	fmt.Fprintln(out)

	for _, t := range planned {
		if err := createSymlink(skillsDir, t.path, out); err != nil {
			return err
		}
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Done. Skills installed under .agents/skills will be visible to:")
	for _, t := range planned {
		fmt.Fprintf(out, "  • %s\n", t.label)
	}
	return nil
}

func promptSetup(cmd *cobra.Command, skillsDirRel string, targets []linkTarget, opts *setupOptions) error {
	claude := new(bool)
	cursor := new(bool)
	kilo := new(bool)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Sqill setup").
				Description(fmt.Sprintf("Initializing %s\n\nChoose which tool directories to symlink into the shared skills folder.", skillsDirRel)),
			huh.NewConfirm().
				Title("Symlink Claude skills?").
				Description("Link .claude/skills into .agents/skills").
				Affirmative("Yes").
				Negative("No").
				Value(claude),
			huh.NewConfirm().
				Title("Symlink Cursor skills?").
				Description("Link .cursor/skills into .agents/skills").
				Affirmative("Yes").
				Negative("No").
				Value(cursor),
			huh.NewConfirm().
				Title("Symlink Kilo skills?").
				Description("Link .kilo/skills into .agents/skills").
				Affirmative("Yes").
				Negative("No").
				Value(kilo),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return fmt.Errorf("setup cancelled")
		}
		return err
	}

	opts.linkClaude = *claude
	opts.linkCursor = *cursor
	opts.linkKilo = *kilo
	return nil
}

func plannedLinks(opts *setupOptions, targets []linkTarget) []linkTarget {
	var out []linkTarget
	for _, t := range targets {
		switch t.name {
		case "claude":
			if opts.linkClaude {
				out = append(out, t)
			}
		case "cursor":
			if opts.linkCursor {
				out = append(out, t)
			}
		case "kilo":
			if opts.linkKilo {
				out = append(out, t)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out
}

func createSymlink(skillsDir, linkPath string, out io.Writer) error {
	info, err := os.Lstat(linkPath)
	if errors.Is(err, os.ErrNotExist) {
		return makeSymlink(skillsDir, linkPath, out)
	}
	if err != nil {
		return fmt.Errorf("check %s: %w", displayPath(linkPath), err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		fmt.Fprintf(out, "• %s already a symlink, skipped\n", displayPath(linkPath))
		return nil
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists and is not a directory", displayPath(linkPath))
	}

	dupes, err := findDuplicates(skillsDir, linkPath)
	if err != nil {
		return err
	}
	if len(dupes) > 0 {
		return fmt.Errorf("cannot link %s — duplicate skill(s): %s. Please de-duplicate and run setup again", displayPath(linkPath), strings.Join(dupes, ", "))
	}

	if err := moveContents(linkPath, skillsDir); err != nil {
		return fmt.Errorf("move contents of %s: %w", displayPath(linkPath), err)
	}
	if err := os.Remove(linkPath); err != nil {
		return fmt.Errorf("remove %s: %w", displayPath(linkPath), err)
	}

	return makeSymlink(skillsDir, linkPath, out)
}

func makeSymlink(skillsDir, linkPath string, out io.Writer) error {
	if err := os.MkdirAll(filepath.Dir(linkPath), 0o755); err != nil {
		return fmt.Errorf("mkdir parent of %s: %w", displayPath(linkPath), err)
	}
	rel, err := filepath.Rel(filepath.Dir(linkPath), skillsDir)
	if err != nil {
		return fmt.Errorf("relative path for %s: %w", displayPath(linkPath), err)
	}
	if err := os.Symlink(rel, linkPath); err != nil {
		return fmt.Errorf("symlink %s -> %s: %w", displayPath(linkPath), rel, err)
	}
	fmt.Fprintf(out, "✓ %s -> %s\n", displayPath(linkPath), displayPath(filepath.Join(filepath.Dir(linkPath), rel)))
	return nil
}

func findDuplicates(skillsDir, linkPath string) ([]string, error) {
	existing, err := subdirNames(linkPath)
	if err != nil {
		return nil, err
	}
	target, err := subdirNames(skillsDir)
	if err != nil {
		return nil, err
	}
	targetSet := make(map[string]struct{}, len(target))
	for _, n := range target {
		targetSet[n] = struct{}{}
	}
	var dupes []string
	for _, n := range existing {
		if _, ok := targetSet[n]; ok {
			dupes = append(dupes, n)
		}
	}
	sort.Strings(dupes)
	return dupes, nil
}

func subdirNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func moveContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if err := os.Rename(s, d); err != nil {
			return fmt.Errorf("rename %s -> %s: %w", displayPath(s), displayPath(d), err)
		}
	}
	return nil
}

func cwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

func displayPath(path string) string {
	cwd := cwdEval()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}
	if rel, err := filepath.Rel(cwd, resolved); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}

func cwdEval() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}
	return dir
}
