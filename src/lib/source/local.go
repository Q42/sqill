package source

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Local struct{}

func NewLocal() *Local { return &Local{} }

func (l *Local) Type() Type { return TypeLocal }

func (l *Local) Fetch(source string, dest string) error {
	src, err := stripFileScheme(source)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("local source %q is not a directory", src)
	}

	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("abs source: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create dest parent: %w", err)
	}

	tmp, err := os.MkdirTemp(filepath.Dir(dest), ".local-")
	if err != nil {
		return fmt.Errorf("temp: %w", err)
	}
	tmpDest := filepath.Join(tmp, "skill")
	if err := os.MkdirAll(tmpDest, 0o755); err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("temp dest: %w", err)
	}

	if err := copyTree(absSrc, tmpDest); err != nil {
		os.RemoveAll(tmp)
		return err
	}

	if err := os.Rename(tmpDest, dest); err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("rename temp: %w", err)
	}
	os.RemoveAll(tmp)
	return nil
}

func stripFileScheme(s string) (string, error) {
	const prefix = "file://"
	if len(s) <= len(prefix) {
		return "", errors.New("invalid file:// source")
	}
	return s[len(prefix):], nil
}

func copyTree(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	for _, entry := range entries {
		s := filepath.Join(src, entry.Name())
		d := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyTree(s, d); err != nil {
				return err
			}
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("stat: %w", err)
		}
		if err := copyFile(s, d, info.Mode()); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
