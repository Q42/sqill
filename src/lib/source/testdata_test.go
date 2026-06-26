package source

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
)

func writeSampleTarGz(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	files := []struct {
		name    string
		content string
	}{
		{"hello.txt", "hi"},
		{filepath.Join("nested", "x.txt"), "yo"},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name: file.name,
			Mode: 0o644,
			Size: int64(len(file.content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(file.content)); err != nil {
			return err
		}
	}
	return nil
}
