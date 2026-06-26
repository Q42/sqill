package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DisplayPath(path string) string {
	cwd := CwdEval()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}
	if rel, err := filepath.Rel(cwd, resolved); err == nil && !strings.HasPrefix(rel, "..") {
		return rel
	}
	return path
}

func CwdEval() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return resolved
	}
	return dir
}

func SafeJoin(root, name string) (string, error) {
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, `\`) {
		return "", fmt.Errorf("unsafe path %q", name)
	}
	parts := strings.Split(filepath.ToSlash(name), "/")
	for _, p := range parts {
		if p == ".." {
			return "", fmt.Errorf("unsafe path %q", name)
		}
	}
	target := filepath.Join(root, name)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return "", fmt.Errorf("unsafe path %q: %w", name, err)
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("unsafe path %q", name)
	}
	return target, nil
}

func SingleRootDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read extract: %w", err)
	}
	if len(entries) != 1 {
		return "", fmt.Errorf("archive must contain a single root directory, got %d", len(entries))
	}
	root := filepath.Join(dir, entries[0].Name())
	if !entries[0].IsDir() {
		return "", fmt.Errorf("archive root %q is not a directory", entries[0].Name())
	}
	return root, nil
}

func SubdirNames(dir string) ([]string, error) {
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

func FindDuplicates(a, b string) ([]string, error) {
	existing, err := SubdirNames(a)
	if err != nil {
		return nil, err
	}
	target, err := SubdirNames(b)
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

func MoveContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		s := filepath.Join(src, e.Name())
		d := filepath.Join(dst, e.Name())
		if err := os.Rename(s, d); err != nil {
			return fmt.Errorf("rename %s -> %s: %w", DisplayPath(s), DisplayPath(d), err)
		}
	}
	return nil
}

func StripGitDirs(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("read dir: %w", err)
	}
	for _, e := range entries {
		if e.Name() == ".git" {
			if err := os.RemoveAll(filepath.Join(root, e.Name())); err != nil {
				return fmt.Errorf("remove .git: %w", err)
			}
			continue
		}
		if e.IsDir() {
			if err := StripGitDirs(filepath.Join(root, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func ValidateName(name string) error {
	if name == "" {
		return errors.New("skill name is empty")
	}
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) || strings.HasPrefix(name, ".") {
		return fmt.Errorf("invalid skill name %q", name)
	}
	return nil
}