package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.dat")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("open empty: %v", err)
	}
	if err := s.Upsert(Entry{Name: "GitHub", Kind: KindToken, Value: "ghp_1234567890abcdef"}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := s.Upsert(Entry{Name: "AWS", Kind: KindAPIKey, Value: "AKIA...", Note: "prod"}); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read vault: %v", err)
	}
	if string(raw) == "" || contains(raw, "ghp_1234567890abcdef") {
		t.Fatalf("vault on disk must be encrypted")
	}

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	list := s2.List()
	if len(list) != 2 || list[0].Name != "AWS" || list[1].Name != "GitHub" {
		t.Fatalf("unexpected list: %+v", list)
	}
	if got := s2.Search("prod"); len(got) != 1 || got[0].Name != "AWS" {
		t.Fatalf("search by note failed: %+v", got)
	}

	id := list[1].ID
	if err := s2.Upsert(Entry{ID: id, Name: "GitHub", Kind: KindToken, Value: "new"}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if e, ok := s2.Get(id); !ok || e.Value != "new" || e.CreatedAt.IsZero() {
		t.Fatalf("update not applied: %+v", e)
	}
	if err := s2.Delete(id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(s2.List()) != 1 {
		t.Fatalf("delete not applied")
	}
}

func TestMasked(t *testing.T) {
	cases := map[string]string{
		"":                     "",
		"short":                "••••••••",
		"12345678":             "12••••••78",
		"ghp_1234567890abcdef": "ghp_••••••••cdef",
	}
	for in, want := range cases {
		if got := (Entry{Value: in}).Masked(); got != want {
			t.Errorf("Masked(%q) = %q, want %q", in, got, want)
		}
	}
}

func contains(b []byte, s string) bool {
	return len(s) > 0 && len(b) >= len(s) && indexOf(string(b), s) >= 0
}

func indexOf(hay, needle string) int {
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
