package installer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sqill/internal/metadata"
	"sqill/internal/registry"
	"sqill/internal/source"
)

type Installer struct {
	registry registry.Provider
	sources  map[source.Type]source.Provider
	store    metadata.Store
	skillsDir string
}

func New(reg registry.Provider, store metadata.Store, skillsDir string) *Installer {
	return &Installer{
		registry:  reg,
		sources:   nil,
		store:     store,
		skillsDir: skillsDir,
	}
}

func defaultSources(skillName string) map[source.Type]source.Provider {
	return map[source.Type]source.Provider{
		source.TypeGit:     source.NewGit(skillName),
		source.TypeLocal:   source.NewLocal(),
		source.TypeArchive: source.NewArchive(),
	}
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

func (i *Installer) Install(name string, force bool) error {
	if err := ValidateName(name); err != nil {
		return err
	}

	entry, err := i.registry.Resolve(name)
	if err != nil {
		return err
	}

	if !force && i.store.IsInstalled(name) {
		return fmt.Errorf("skill %q already installed (use --force to overwrite)", name)
	}

	stype, err := source.Detect(entry.Source)
	if err != nil {
		return err
	}
	prov, ok := i.getSources(name)[stype]
	if !ok {
		return fmt.Errorf("no provider for source type %q", stype)
	}

	if err := os.MkdirAll(i.skillsDir, 0o755); err != nil {
		return fmt.Errorf("create skills dir: %w", err)
	}

	target := filepath.Join(i.skillsDir, name)
	if force {
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("clear target: %w", err)
		}
	}

	if err := prov.Fetch(entry.Source, target); err != nil {
		return err
	}

	manifest, err := metadata.LoadManifest(target)
	if err != nil {
		os.RemoveAll(target)
		return fmt.Errorf("validate manifest: %w", err)
	}
	if manifest.Name != name {
		os.RemoveAll(target)
		return fmt.Errorf("manifest name %q does not match requested %q", manifest.Name, name)
	}

	now := metadata.Now()
	if err := i.store.Add(name, metadata.InstalledEntry{
		Version:     manifest.Version,
		Source:      entry.Source,
		InstalledAt: now,
		Description: manifest.Description,
	}); err != nil {
		os.RemoveAll(target)
		return fmt.Errorf("write metadata: %w", err)
	}
	return nil
}

func (i *Installer) SkillDir(name string) string {
	return filepath.Join(i.skillsDir, name)
}

func (i *Installer) getSources(skillName string) map[source.Type]source.Provider {
	if i.sources == nil {
		i.sources = defaultSources(skillName)
		return i.sources
	}
	i.sources[source.TypeGit] = source.NewGit(skillName)
	return i.sources
}
