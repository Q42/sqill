package init

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

	"sqill/src/lib/metadata"
	"sqill/src/lib/utils"
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

type Options struct {
	SkillsDir  string
	LinkClaude bool
	LinkCursor bool
	LinkKilo   bool
	Yes        bool
}

func NewCmd() *cobra.Command {
	opts := &Options{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize .agents/skills and optionally create tool symlinks",
		Long:  "Creates .agents/skills/sqill.json if missing, then optionally symlinks .claude/skills, .cursor/skills, and .kilo/skills into it.",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cmd.Flags().GetString("skills-dir")
			if err != nil {
				return err
			}
			opts.SkillsDir = dir
			return run(cmd, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.LinkClaude, "link-claude", false, "Symlink .claude/skills into .agents/skills")
	cmd.Flags().BoolVar(&opts.LinkCursor, "link-cursor", false, "Symlink .cursor/skills into .agents/skills")
	cmd.Flags().BoolVar(&opts.LinkKilo, "link-kilo", false, "Symlink .kilo/skills into .agents/skills")
	cmd.Flags().BoolVar(&opts.Yes, "yes", false, "Skip prompts (no symlinks by default)")

	return cmd
}

func run(cmd *cobra.Command, opts *Options) error {
	out := cmd.OutOrStderr()

	skillsDir, err := filepath.Abs(opts.SkillsDir)
	if err != nil {
		return fmt.Errorf("resolve skills dir: %w", err)
	}
	skillsDirRel, _ := filepath.Rel(utils.CwdEval(), skillsDir)

	statePath := filepath.Join(skillsDir, metadata.StateFileName)
	stateExists := false
	if _, err := os.Stat(statePath); err == nil {
		stateExists = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check state file: %w", err)
	}

	if stateExists {
		fmt.Fprintf(out, "✓ %s is already initialized (%s exists)\n", utils.DisplayPath(skillsDir), metadata.StateFileName)
	}

	baseDir := filepath.Dir(filepath.Dir(skillsDir))
	targets := defaultLinkTargets(baseDir)

	flagsProvided := opts.LinkClaude || opts.LinkCursor || opts.LinkKilo
	interactive := !flagsProvided && !opts.Yes && !stateExists

	if interactive {
		if err := prompt(cmd, skillsDirRel, targets, opts); err != nil {
			return err
		}
	} else if opts.Yes && !flagsProvided {
		opts.LinkClaude = false
		opts.LinkCursor = false
		opts.LinkKilo = false
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
		fmt.Fprintf(out, "✓ Created %s\n", utils.DisplayPath(statePath))
	}

	planned := plannedLinks(opts, targets)
	if len(planned) == 0 {
		fmt.Fprintln(out, "• No symlinks requested")
		return nil
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Setting up symlinks:")
	for _, t := range planned {
		fmt.Fprintf(out, "  → %s\n", utils.DisplayPath(t.path))
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

	if err := metadata.SyncGitignore(skillsDir); err != nil {
		return fmt.Errorf("sync gitignore: %w", err)
	}
	return nil
}

func prompt(cmd *cobra.Command, skillsDirRel string, targets []linkTarget, opts *Options) error {
	claude := new(bool)
	cursor := new(bool)
	kilo := new(bool)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Sqill init").
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
			return fmt.Errorf("init cancelled")
		}
		return err
	}

	opts.LinkClaude = *claude
	opts.LinkCursor = *cursor
	opts.LinkKilo = *kilo
	return nil
}

func plannedLinks(opts *Options, targets []linkTarget) []linkTarget {
	var out []linkTarget
	for _, t := range targets {
		switch t.name {
		case "claude":
			if opts.LinkClaude {
				out = append(out, t)
			}
		case "cursor":
			if opts.LinkCursor {
				out = append(out, t)
			}
		case "kilo":
			if opts.LinkKilo {
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
		return fmt.Errorf("check %s: %w", utils.DisplayPath(linkPath), err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		fmt.Fprintf(out, "• %s already a symlink, skipped\n", utils.DisplayPath(linkPath))
		return nil
	}
	if !info.IsDir() {
		return fmt.Errorf("%s exists and is not a directory", utils.DisplayPath(linkPath))
	}

	dupes, err := utils.FindDuplicates(skillsDir, linkPath)
	if err != nil {
		return err
	}
	if len(dupes) > 0 {
		return fmt.Errorf("cannot link %s — duplicate skill(s): %s. Please de-duplicate and run init again", utils.DisplayPath(linkPath), strings.Join(dupes, ", "))
	}

	if err := utils.MoveContents(linkPath, skillsDir); err != nil {
		return fmt.Errorf("move contents of %s: %w", utils.DisplayPath(linkPath), err)
	}
	if err := os.Remove(linkPath); err != nil {
		return fmt.Errorf("remove %s: %w", utils.DisplayPath(linkPath), err)
	}

	return makeSymlink(skillsDir, linkPath, out)
}

func makeSymlink(skillsDir, linkPath string, out io.Writer) error {
	if err := os.MkdirAll(filepath.Dir(linkPath), 0o755); err != nil {
		return fmt.Errorf("mkdir parent of %s: %w", utils.DisplayPath(linkPath), err)
	}
	rel, err := filepath.Rel(filepath.Dir(linkPath), skillsDir)
	if err != nil {
		return fmt.Errorf("relative path for %s: %w", utils.DisplayPath(linkPath), err)
	}
	if err := os.Symlink(rel, linkPath); err != nil {
		return fmt.Errorf("symlink %s -> %s: %w", utils.DisplayPath(linkPath), rel, err)
	}
	fmt.Fprintf(out, "✓ %s -> %s\n", utils.DisplayPath(linkPath), utils.DisplayPath(filepath.Join(filepath.Dir(linkPath), rel)))
	return nil
}