package installer

import (
	"fmt"
	"os"

	"sqill/src/lib/metadata"
	"sqill/src/lib/source"
	"sqill/src/lib/utils"
)

func (i *Installer) Update(name string) error {
	if err := utils.ValidateName(name); err != nil {
		return err
	}

	entry, err := i.store.Get(name)
	if err != nil {
		return err
	}

	stype, err := source.Detect(entry.Source)
	if err != nil {
		return err
	}
	prov, ok := i.getSources()[stype]
	if !ok {
		return fmt.Errorf("no provider for source type %q", stype)
	}

	target := i.SkillDir(name)
	tmp := target + ".new"
	if err := os.RemoveAll(tmp); err != nil {
		return fmt.Errorf("clear tmp: %w", err)
	}

	if err := prov.Fetch(entry.Source, tmp); err != nil {
		os.RemoveAll(tmp)
		return err
	}

	manifest, err := metadata.LoadManifest(tmp)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("validate manifest: %w", err)
	}
	if manifest.Name != name {
		os.RemoveAll(tmp)
		return fmt.Errorf("manifest name %q does not match requested %q", manifest.Name, name)
	}

	if err := os.RemoveAll(target); err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("remove old: %w", err)
	}
	if err := os.Rename(tmp, target); err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("rename new: %w", err)
	}

	if err := i.store.Add(name, metadata.InstalledEntry{
		Version:     manifest.Version,
		Source:      entry.Source,
		InstalledAt: metadata.Now(),
		Description: manifest.Description,
	}); err != nil {
		return fmt.Errorf("update metadata: %w", err)
	}
	return nil
}