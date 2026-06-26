package source

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"sqill/src/lib/utils"
)

type Archive struct {
	Client *http.Client
}

func NewArchive() *Archive {
	return &Archive{Client: &http.Client{Timeout: 5 * time.Minute}}
}

func (a *Archive) Type() Type { return TypeArchive }

func (a *Archive) Fetch(source string, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create dest parent: %w", err)
	}

	client := a.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Minute}
	}

	resp, err := client.Get(source)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download: status %d", resp.StatusCode)
	}

	tmp, err := os.MkdirTemp(filepath.Dir(dest), ".archive-")
	if err != nil {
		return fmt.Errorf("temp: %w", err)
	}
	archivePath := filepath.Join(tmp, "skill.tar.gz")
	out, err := os.Create(archivePath)
	if err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("create archive: %w", err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.RemoveAll(tmp)
		return fmt.Errorf("save archive: %w", err)
	}
	if err := out.Close(); err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("close archive: %w", err)
	}

	tmpExtract := filepath.Join(tmp, "extract")
	if err := os.MkdirAll(tmpExtract, 0o755); err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("mkdir extract: %w", err)
	}
	if err := extractTarGz(archivePath, tmpExtract); err != nil {
		os.RemoveAll(tmp)
		return err
	}

	root, err := utils.SingleRootDir(tmpExtract)
	if err != nil {
		os.RemoveAll(tmp)
		return err
	}

	if err := os.Rename(root, dest); err != nil {
		os.RemoveAll(tmp)
		return fmt.Errorf("rename: %w", err)
	}
	os.RemoveAll(tmp)
	return nil
}

func extractTarGz(archivePath, dest string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar next: %w", err)
		}

		target, err := utils.SafeJoin(dest, hdr.Name)
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("mkdir %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return fmt.Errorf("mkdir parent %s: %w", target, err)
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0o777)
			if err != nil {
				return fmt.Errorf("create %s: %w", target, err)
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return fmt.Errorf("write %s: %w", target, err)
			}
			if err := out.Close(); err != nil {
				return fmt.Errorf("close %s: %w", target, err)
			}
		case tar.TypeSymlink, tar.TypeLink:
			_ = os.RemoveAll(target)
			if hdr.Typeflag == tar.TypeSymlink {
				if err := os.Symlink(hdr.Linkname, target); err != nil {
					return fmt.Errorf("symlink %s: %w", target, err)
				}
			}
		default:
		}
	}
	return nil
}