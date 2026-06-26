package source

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Git struct{}

func NewGit() *Git { return &Git{} }

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

	os.RemoveAll(dest)
	if err := os.Rename(tmp, dest); err != nil {
		return fmt.Errorf("move to dest: %w", err)
	}
	return nil
}
