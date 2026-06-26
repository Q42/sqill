package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"sqill/src/lib/metadata"
	"sqill/src/lib/registry"
	"sqill/src/lib/source"
	"sqill/src/lib/utils"
)

type Installer struct {
	registry  registry.Provider
	sources   map[source.Type]source.Provider
	store     metadata.Store
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

func defaultSources() map[source.Type]source.Provider {
	return map[source.Type]source.Provider{
		source.TypeGit:     source.NewGit(),
		source.TypeLocal:   source.NewLocal(),
		source.TypeArchive: source.NewArchive(),
	}
}

func (i *Installer) Install(name string, force bool) error {
	if err := utils.ValidateName(name); err != nil {
		return err
	}

	entry, err := i.registry.Resolve(name)
	if err != nil {
		return err
	}

	if !force && i.store.IsInstalled(name) {
		return fmt.Errorf("skill %q already installed (use --force to overwrite)", name)
	}

	target := filepath.Join(i.skillsDir, name)
	if force {
		if err := os.RemoveAll(target); err != nil {
			return fmt.Errorf("clear target: %w", err)
		}
	}

	manifest, err := i.fetchAndStage(name, entry.Source, target)
	if err != nil {
		return err
	}

	if err := i.store.Add(name, metadata.InstalledEntry{
		Version:     manifest.Version,
		Source:      entry.Source,
		InstalledAt: metadata.Now(),
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

func pickManifestDir(target, skillName string) (string, error) {
	candidates := []string{
		target,
		filepath.Join(target, skillName),
		filepath.Join(target, "skill"),
		filepath.Join(target, "sqill"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "sqill.json")); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("read manifest: %s: no such file or directory", filepath.Join(target, "sqill.json"))
}

func (i *Installer) getSources() map[source.Type]source.Provider {
	if i.sources == nil {
		i.sources = defaultSources()
		return i.sources
	}
	i.sources[source.TypeGit] = source.NewGit()
	return i.sources
}
