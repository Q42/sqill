package source

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Git struct {
	SkillName string
}

func NewGit(skillName string) *Git { return &Git{SkillName: skillName} }

func (g *Git) Type() Type { return TypeGit }

func (g *Git) Fetch(source string, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create dest parent: %w", err)
	}

	tmp, err := os.MkdirTemp(filepath.Dir(dest), ".git-clone-")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	cmd := exec.Command("git", "clone", source, tmp)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone %q: %w", source, err)
	}

	root := pickSubfolder(tmp, g.SkillName)

	os.RemoveAll(dest)
	if err := os.Rename(root, dest); err != nil {
		return fmt.Errorf("move to dest: %w", err)
	}
	return nil
}

func pickSubfolder(cloneDir, skillName string) string {
	candidates := []string{skillName, "skill"}
	for _, name := range candidates {
		if name == "" {
			continue
		}
		path := filepath.Join(cloneDir, name)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
	}
	return cloneDir
}
