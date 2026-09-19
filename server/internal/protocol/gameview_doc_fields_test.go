package protocol

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

// TestProtocolDocGameViewFieldsMatchStruct guards against the doc drift
// reported in #805: docs/protocol.md's "GameView schema (summary)"
// bullet must name exactly the JSON keys GameView actually serialises,
// no more and no fewer. When this test fails, regenerate the summary
// bullet in docs/protocol.md from the struct's `json` tags below —
// never edit GameView to fit a stale doc.
func TestProtocolDocGameViewFieldsMatchStruct(t *testing.T) {
	structFields := gameViewJSONFieldNames(t)
	docFields := documentedGameViewFieldNames(t)

	sort.Strings(structFields)
	sort.Strings(docFields)

	missing := stringsNotIn(structFields, docFields) // on the struct, absent from the doc
	extra := stringsNotIn(docFields, structFields)   // in the doc, absent from the struct

	if len(missing) > 0 {
		t.Errorf("docs/protocol.md GameView summary is missing fields the struct has: %v", missing)
	}
	if len(extra) > 0 {
		t.Errorf("docs/protocol.md GameView summary names fields the struct doesn't have (renamed or removed?): %v", extra)
	}
}

// gameViewJSONFieldNames reflects over GameView and returns every
// exported field's wire name, taken from its `json` struct tag. Fields
// tagged "-" are skipped, as are unexported fields such as
// legalBySeat, which encoding/json never touches and which carry no
// json tag to read.
func gameViewJSONFieldNames(t *testing.T) []string {
	t.Helper()

	typ := reflect.TypeOf(GameView{})
	names := make([]string, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.PkgPath != "" {
			continue // unexported
		}
		tag := f.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	return names
}

// gameViewSummaryPattern matches the
// "- **GameView**: `{ id, state, ... }`" bullet under the "GameView
// schema (summary)" heading in docs/protocol.md, capturing the
// braced field list.
var gameViewSummaryPattern = regexp.MustCompile("(?m)^-\\s+\\*\\*GameView\\*\\*:\\s+`\\{([^}]*)\\}`")

// documentedGameViewFieldNames parses that bullet and returns the field
// names it lists, with the trailing `[]` (slice) and `?` (omitempty)
// markers stripped so they line up with the struct's bare JSON names.
func documentedGameViewFieldNames(t *testing.T) []string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not locate this test file")
	}
	// This file lives at server/internal/protocol/<name>.go, so the
	// repo root is three directories up.
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	docPath := filepath.Join(repoRoot, "docs", "protocol.md")

	raw, err := os.ReadFile(docPath)
	if err != nil {
		t.Fatalf("reading %s: %v", docPath, err)
	}

	match := gameViewSummaryPattern.FindStringSubmatch(string(raw))
	if match == nil {
		t.Fatalf("could not find the \"- **GameView**: `{ ... }`\" summary bullet in %s", docPath)
	}

	var names []string
	for _, part := range strings.Split(match[1], ",") {
		part = strings.TrimSpace(part)
		part = strings.TrimSuffix(part, "?")
		part = strings.TrimSuffix(part, "[]")
		if part == "" {
			continue
		}
		names = append(names, part)
	}
	return names
}

// stringsNotIn returns the elements of a that are absent from b.
func stringsNotIn(a, b []string) []string {
	inB := make(map[string]bool, len(b))
	for _, s := range b {
		inB[s] = true
	}
	var out []string
	for _, s := range a {
		if !inB[s] {
			out = append(out, s)
		}
	}
	return out
}
