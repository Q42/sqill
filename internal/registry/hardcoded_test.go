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

func TestHardcodedSearch(t *testing.T) {
	r := NewHardcoded()
	all := r.Search("")
	if len(all) != len(defaultRegistry) {
		t.Fatalf("expected %d entries, got %d", len(defaultRegistry), len(all))
	}
	git := r.Search("sRegressor")
	if len(git) == 0 {
		t.Fatal("expected at least one match")
	}
	for _, e := range git {
		if !contains(e.Name, "sRegressor") && !contains(e.Description, "sRegressor") {
			t.Fatalf("unexpected match %+v", e)
		}
	}
}

func TestHardcodedAll(t *testing.T) {
	r := NewHardcoded()
	all := r.All()
	if len(all) != len(defaultRegistry) {
		t.Fatalf("got %d", len(all))
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
