package registry

import "testing"

func TestHardcodedResolve(t *testing.T) {
	r := NewHardcoded()
	e, err := r.Resolve("sRegressor")
	if err != nil {
		t.Fatal(err)
	}
	if e.Source == "" {
		t.Fatal("expected source")
	}
	if _, err := r.Resolve("nope"); err == nil {
		t.Fatal("expected error")
	}
}

func TestHardcodedAll(t *testing.T) {
	r := NewHardcoded()
	all := r.All()
	if len(all) != len(defaultRegistry) {
		t.Fatalf("got %d", len(all))
	}
}
