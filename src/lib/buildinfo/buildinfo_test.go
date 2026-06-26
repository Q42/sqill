package buildinfo

import "testing"

func TestVersionDefault(t *testing.T) {
	if Version == "" {
		t.Fatal("expected non-empty default Version")
	}
}
