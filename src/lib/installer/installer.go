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

	stype, err := source.Detect(entry.Source)
	if err != nil {
		return err
	}
	prov, ok := i.getSources()[stype]
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

	manifestDir, err := pickManifestDir(target, name)
	if err != nil {
		os.RemoveAll(target)
		return fmt.Errorf("validate manifest: %w", err)
	}
	if manifestDir != target {
		if err := utils.MoveContents(manifestDir, target); err != nil {
			os.RemoveAll(target)
			return fmt.Errorf("flatten subdir: %w", err)
		}
		os.RemoveAll(manifestDir)
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