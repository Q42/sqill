package installer

import (
	"fmt"
	"os"

	"sqill/src/lib/utils"
)

func (i *Installer) Remove(name string) error {
	if err := utils.ValidateName(name); err != nil {
		return err
	}
	if !i.store.IsInstalled(name) {
		return fmt.Errorf("skill %q not installed", name)
	}

	target := i.SkillDir(name)
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("remove directory: %w", err)
	}

	if err := i.store.Remove(name); err != nil {
		return fmt.Errorf("remove metadata: %w", err)
	}
	return nil
}