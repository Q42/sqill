package upgrader

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultRepo   = "Q42/sqill"
	DefaultBinary = "sqill"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Options struct {
	Repo    string
	Binary  string
	Version string
	Force   bool
	Client  HTTPDoer
	GOOS    string
	GOARCH  string
	ExePath string
}

type Result struct {
	From    string
	To      string
	Skipped bool
}

func (o *Options) defaults() {
	if o.Repo == "" {
		o.Repo = DefaultRepo
	}
	if o.Binary == "" {
		o.Binary = DefaultBinary
	}
	if o.Client == nil {
		o.Client = &http.Client{Timeout: 5 * time.Minute}
	}
	if o.GOOS == "" {
		o.GOOS = runtime.GOOS
	}
	if o.GOARCH == "" {
		o.GOARCH = runtime.GOARCH
	}
	if o.ExePath == "" {
		exe, err := os.Executable()
		if err == nil {
			o.ExePath = exe
		}
	}
}

func AssetName(binary, goos, goarch string) string {
	return fmt.Sprintf("%s_%s_%s.tar.gz", binary, goos, goarch)
}

func ReleaseURL(repo, tag, asset string) string {
	return fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, tag, asset)
}

func LatestTag(client HTTPDoer, repo string) (string, error) {
	if repo == "" {
		repo = DefaultRepo
	}
	return LatestTagAtURL(client, fmt.Sprintf("https://github.com/%s/releases/latest", repo))
}

func LatestTagAtURL(client HTTPDoer, url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("resolve latest: %w", err)
	}
	defer resp.Body.Close()
	if loc := resp.Header.Get("Location"); loc != "" {
		tag := tagFromURL(loc)
		if tag != "" {
			return tag, nil
		}
	}
	return "", fmt.Errorf("resolve latest: status %d (no redirect)", resp.StatusCode)
}

func tagFromURL(rawURL string) string {
	const marker = "/releases/tag/"
	i := strings.Index(rawURL, marker)
	if i < 0 {
		return ""
	}
	return rawURL[i+len(marker):]
}

func normalizeTag(tag string) string {
	return strings.TrimPrefix(tag, "v")
}

func Run(opts Options) (*Result, error) {
	opts.defaults()
	if opts.ExePath == "" {
		return nil, errors.New("could not determine current executable path")
	}

	current := normalizeTag(opts.Version)
	latest, err := LatestTag(opts.Client, opts.Repo)
	if err != nil {
		return nil, err
	}
	latestNorm := normalizeTag(latest)

	if !opts.Force && current != "" && current != "dev" && current == latestNorm {
		return &Result{From: current, To: latest, Skipped: true}, nil
	}

	url := ReleaseURL(opts.Repo, latest, AssetName(opts.Binary, opts.GOOS, opts.GOARCH))

	tmp, err := os.MkdirTemp(filepath.Dir(opts.ExePath), ".sqill-upgrade-")
	if err != nil {
		return nil, fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	archivePath := filepath.Join(tmp, "sqill.tar.gz")
	if err := download(opts.Client, url, archivePath); err != nil {
		return nil, err
	}

	newBinary := filepath.Join(tmp, opts.Binary)
	if err := extractBinary(archivePath, newBinary); err != nil {
		return nil, err
	}
	if err := os.Chmod(newBinary, 0o755); err != nil {
		return nil, fmt.Errorf("chmod: %w", err)
	}

	if err := replaceExecutable(opts.ExePath, newBinary); err != nil {
		return nil, err
	}

	return &Result{From: current, To: latest}, nil
}

func download(client HTTPDoer, url, dest string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("create %s: %w", dest, err)
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(dest)
		return fmt.Errorf("save: %w", err)
	}
	if err := out.Close(); err != nil {
		os.Remove(dest)
		return fmt.Errorf("close: %w", err)
	}
	return nil
}

func extractBinary(archivePath, dest string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		base := filepath.Base(hdr.Name)
		if base != filepath.Base(dest) {
			continue
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return fmt.Errorf("create %s: %w", dest, err)
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return fmt.Errorf("write: %w", err)
		}
		return out.Close()
	}
	return fmt.Errorf("binary %q not found in archive", filepath.Base(dest))
}

func replaceExecutable(current, newPath string) error {
	if err := os.RemoveAll(current); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove old binary: %w", err)
	}
	if err := os.Rename(newPath, current); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}
