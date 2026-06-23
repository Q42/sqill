package source

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	gogit "github.com/go-git/go-git/v5"
)

type Git struct{}

func NewGit() *Git { return &Git{} }

func (g *Git) Type() Type { return TypeGit }

func (g *Git) Fetch(source string, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create dest parent: %w", err)
	}

	tmp, err := os.MkdirTemp(filepath.Dir(dest), ".git-")
	if err != nil {
		return fmt.Errorf("temp: %w", err)
	}
	tmpClone := filepath.Join(tmp, "repo")

	_, err = gogit.PlainCloneContext(context.Background(), tmpClone, false, &gogit.CloneOptions{
		URL:      source,
		Progress: nil,
	})
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("git clone %q: %w", source, err)
	}

	if err := os.Rename(tmpClone, dest); err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("rename clone: %w", err)
	}
	os.RemoveAll(tmp)
	return nil
}
