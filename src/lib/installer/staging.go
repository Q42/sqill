package installer

import (
	"fmt"
	"os"

	"sqill/src/lib/metadata"
	"sqill/src/lib/source"
	"sqill/src/lib/utils"
)

func (i *Installer) fetchAndStage(name, src, dest string) (metadata.Manifest, error) {
	stype, err := source.Detect(src)
	if err != nil {
		return metadata.Manifest{}, err
	}
	prov, ok := i.getSources()[stype]
	if !ok {
		return metadata.Manifest{}, fmt.Errorf("no provider for source type %q", stype)
	}

	if err := os.MkdirAll(i.skillsDir, 0o755); err != nil {
		return metadata.Manifest{}, fmt.Errorf("create skills dir: %w", err)
	}

	if err := prov.Fetch(src, dest); err != nil {
		os.RemoveAll(dest)
		return metadata.Manifest{}, err
	}

	manifestDir, err := pickManifestDir(dest, name)
	if err != nil {
		os.RemoveAll(dest)
		return metadata.Manifest{}, fmt.Errorf("validate manifest: %w", err)
	}
	if manifestDir != dest {
		if err := utils.MoveContents(manifestDir, dest); err != nil {
			os.RemoveAll(dest)
			return metadata.Manifest{}, fmt.Errorf("flatten subdir: %w", err)
		}
		os.RemoveAll(manifestDir)
	}

	manifest, err := metadata.LoadManifest(dest)
	if err != nil {
		os.RemoveAll(dest)
		return metadata.Manifest{}, fmt.Errorf("validate manifest: %w", err)
	}
	if manifest.Name != name {
		os.RemoveAll(dest)
		return metadata.Manifest{}, fmt.Errorf("manifest name %q does not match requested %q", manifest.Name, name)
	}
	return manifest, nil
}
