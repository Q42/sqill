package source

import (
	"errors"
	"fmt"
	"strings"
)

type Type string

const (
	TypeGit     Type = "git"
	TypeLocal   Type = "local"
	TypeArchive Type = "archive"
)

type Provider interface {
	Fetch(source string, dest string) error
	Type() Type
}

func Detect(source string) (Type, error) {
	if source == "" {
		return "", errors.New("source is empty")
	}
	switch {
	case strings.HasPrefix(source, "git@"),
		strings.HasSuffix(source, ".git"),
		strings.HasPrefix(source, "https://") && strings.Contains(source, ".git"),
		strings.HasPrefix(source, "http://") && strings.Contains(source, ".git"):
		return TypeGit, nil
	case strings.HasPrefix(source, "file://"):
		return TypeLocal, nil
	case strings.HasPrefix(source, "https://") && (strings.HasSuffix(source, ".tar.gz") || strings.HasSuffix(source, ".tgz")),
		strings.HasPrefix(source, "http://") && (strings.HasSuffix(source, ".tar.gz") || strings.HasSuffix(source, ".tgz")):
		return TypeArchive, nil
	}
	return "", fmt.Errorf("unsupported source: %q", source)
}

func ForSourceType(t Type, p Provider) error {
	if p.Type() != t {
		return fmt.Errorf("provider type %q does not match source type %q", p.Type(), t)
	}
	return nil
}
