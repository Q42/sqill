package installer

import (
	"fmt"
	"os"

	"sqill/src/lib/metadata"
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

	target := i.SkillDir(name)
	tmp := target + ".new"
	if err := os.RemoveAll(tmp); err != nil {
		return fmt.Errorf("clear tmp: %w", err)
	}

	manifest, err := i.fetchAndStage(name, entry.Source, tmp)
	if err != nil {
		os.RemoveAll(tmp)
		return err
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
