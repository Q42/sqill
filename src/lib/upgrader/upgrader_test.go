package upgrader

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestAssetName(t *testing.T) {
	got := AssetName("sqill", "linux", "amd64")
	want := "sqill_linux_amd64.tar.gz"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestReleaseURL(t *testing.T) {
	got := ReleaseURL("Q42/sqill", "v0.1.0", "sqill_darwin_arm64.tar.gz")
	want := "https://github.com/Q42/sqill/releases/download/v0.1.0/sqill_darwin_arm64.tar.gz"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTagFromURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"https://github.com/Q42/sqill/releases/tag/v0.1.0", "v0.1.0"},
		{"https://github.com/Q42/sqill/releases/tag/v1.2.3-rc1", "v1.2.3-rc1"},
		{"https://example.com/no/tag/here", ""},
	}
	for _, c := range cases {
		if got := tagFromURL(c.in); got != c.want {
			t.Errorf("tagFromURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeTag(t *testing.T) {
	cases := []struct{ in, want string }{
		{"v0.1.0", "0.1.0"},
		{"0.1.0", "0.1.0"},
		{"v", ""},
	}
	for _, c := range cases {
		if got := normalizeTag(c.in); got != c.want {
			t.Errorf("normalizeTag(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLatestTag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases/latest" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Location", "https://github.example.com/Q42/sqill/releases/tag/v9.9.9")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	got, err := LatestTag(client, "Q42/sqill")
	if err != nil {
		t.Fatalf("LatestTag: %v", err)
	}
	if got != "v9.9.9" {
		t.Errorf("got %q, want v9.9.9", got)
	}
}

func TestLatestTagFallsBackToRedirectClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "https://example.com/Q42/sqill/releases/tag/v1.2.3")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	got, err := LatestTag(client, "Q42/sqill")
	if err != nil {
		t.Fatalf("LatestTag: %v", err)
	}
	if got != "v1.2.3" {
		t.Errorf("got %q, want v1.2.3", got)
	}
}

func TestExtractBinary(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "in.tar.gz")
	dest := filepath.Join(dir, "sqill")

	writeArchive(t, archive, map[string]string{
		"sqill": "fake binary contents",
		"README.md": "noise",
	})

	if err := extractBinary(archive, dest); err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "fake binary contents" {
		t.Errorf("got %q", string(got))
	}
}

func TestExtractBinaryMissing(t *testing.T) {
	dir := t.TempDir()
	archive := filepath.Join(dir, "in.tar.gz")
	dest := filepath.Join(dir, "sqill")

	writeArchive(t, archive, map[string]string{
		"other": "x",
	})

	if err := extractBinary(archive, dest); err == nil {
		t.Fatal("expected error when binary not in archive")
	}
}

func TestReplaceExecutable(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "sqill")
	if err := os.WriteFile(current, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	newPath := filepath.Join(dir, "sqill.new")
	if err := os.WriteFile(newPath, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceExecutable(current, newPath); err != nil {
		t.Fatalf("replaceExecutable: %v", err)
	}
	got, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("got %q", string(got))
	}
	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		t.Errorf("expected newPath removed, got %v", err)
	}
}

func writeArchive(t *testing.T, path string, files map[string]string) {
	t.Helper()
	out, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	gz := gzip.NewWriter(out)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()
	for name, body := range files {
		hdr := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(tw, body); err != nil {
			t.Fatal(err)
		}
	}
}
