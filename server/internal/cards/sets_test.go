package cards

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestIndexLoadSets(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sets.json")
	body := `{"object":"list","has_more":false,"data":[
		{"object":"set","code":"tmt","name":"Teenage Mutant Ninja Turtles","set_type":"expansion","released_at":"2026-03-06"},
		{"object":"set","code":"PZA","name":"TMNT Source Material","set_type":"masterpiece","parent_set_code":"TMT","released_at":"2026-03-07"}]}`
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	idx := NewIndex()
	if idx.SetsLoaded() {
		t.Fatal("a fresh index must have no set table")
	}
	n, err := idx.LoadSets(p)
	if err != nil || n != 2 {
		t.Fatalf("LoadSets = %d, %v", n, err)
	}
	if !idx.SetsLoaded() {
		t.Fatal("SetsLoaded = false after a load")
	}
	got, ok := idx.Set("pza")
	want := SetInfo{Code: "pza", Name: "TMNT Source Material", SetType: "masterpiece", ParentSetCode: "tmt", ReleasedAt: "2026-03-07"}
	if !ok || got != want {
		t.Fatalf("Set(pza) = %+v, %v; want %+v", got, ok, want)
	}
}

func TestIndexLoadSetsMissingOrBrokenFile(t *testing.T) {
	idx := NewIndex()
	if _, err := idx.LoadSets(filepath.Join(t.TempDir(), "none.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a missing file: err = %v, want os.ErrNotExist", err)
	}
	for _, body := range []string{"<html>maintenance</html>", `{"object":"error","status":503}`} {
		p := filepath.Join(t.TempDir(), "sets.json")
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := idx.LoadSets(p); err == nil {
			t.Errorf("LoadSets(%q) = nil error", body)
		}
	}
	if idx.SetsLoaded() {
		t.Error("a failed load left a set table")
	}
}
