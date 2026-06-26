package runtime

import (
	"fmt"
	"os"
	"path/filepath"

	"sqill/src/lib/installer"
	"sqill/src/lib/metadata"
	"sqill/src/lib/registry"
)

type Runtime struct {
	SkillsDir string
	Store     metadata.Store
	Inst      *installer.Installer
	Reg       registry.Provider
}

func New(skillsDir string) (*Runtime, error) {
	abs, err := filepath.Abs(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("resolve skills dir: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("create skills dir: %w", err)
	}
	store, err := metadata.NewFileStore(abs)
	if err != nil {
		return nil, err
	}
	reg := registry.NewHardcoded()
	return &Runtime{
		SkillsDir: abs,
		Store:     store,
		Inst:      installer.New(reg, store, abs),
		Reg:       reg,
	}, nil
}