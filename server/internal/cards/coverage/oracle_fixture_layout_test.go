package coverage

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// oracle_fixture_layout_test.go — the per-card fixture layout (#1542)
// tested without the Scryfall dump: the loader, the missing-file
// guard, the stale-file diff and the scoped writer. The dump-gated
// TestOracleFixtureIsCurrent runs the same functions against the real
// catalog nightly; these pin what they do on a table small enough to
// read.

const (
	fxA = "0000000a-0000-4000-8000-000000000000"
	fxB = "0000000b-0000-4000-8000-000000000000"
	fxC = "0000000c-0000-4000-8000-000000000000"
	fxX = "0000000e-0000-4000-8000-000000000000"
	fxY = "0000000f-0000-4000-8000-000000000000"
)

func mustEncode(t *testing.T, c OracleCard) []byte {
	t.Helper()
	b, err := EncodeOracleCard(c)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// The migration contract: a file holds exactly the row the old single
// file held — one line, HTML characters unescaped — plus a newline, and
// loads back to the same card.
func TestEncodeOracleCardIsTheOldRowAndRoundTrips(t *testing.T) {
	c := OracleCard{Name: "Up // Down", Faces: []OracleFace{
		{Name: "Up", Text: "Draw a card.\n{T}: Add {C}."},
		{Name: "Down", Text: "Target creature gets -2/-2 & can't block <this turn>."},
	}}
	b := mustEncode(t, c)
	want := `{"name":"Up // Down","faces":[{"name":"Up","text":"Draw a card.\n{T}: Add {C}."},{"name":"Down","text":"Target creature gets -2/-2 & can't block <this turn>."}]}` + "\n"
	if string(b) != want {
		t.Fatalf("EncodeOracleCard =\n%s\nwant\n%s", b, want)
	}
	got, raw, err := LoadOracleFixtureFS(fstest.MapFS{"d/" + fxA + ".json": {Data: b}}, "d")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, map[string]OracleCard{fxA: c}) || !bytes.Equal(raw[fxA], b) {
		t.Fatalf("round trip: got %#v", got)
	}
}

// A file the generator would never write is an error, not a skip: a
// skipped file is a card whose text silently stops being checked.
func TestLoadOracleFixtureRejectsStrayFiles(t *testing.T) {
	good := mustEncode(t, OracleCard{Name: "Good", Text: "x"})
	for name, extra := range map[string]*fstest.MapFile{
		"README.md":                    {Data: []byte("hi")},
		fxB + "#1.json":                {Data: good},
		strings.ToUpper(fxB) + ".json": {Data: good},
		"not-a-uuid.json":              {Data: good},
		fxB + ".json.orig":             {Data: good},
	} {
		fsys := fstest.MapFS{"d/" + fxA + ".json": {Data: good}, "d/" + name: extra}
		if _, _, err := LoadOracleFixtureFS(fsys, "d"); err == nil {
			t.Errorf("a file named %q loaded without error", name)
		}
	}
	bad := fstest.MapFS{"d/" + fxA + ".json": {Data: []byte("{not json")}}
	if _, _, err := LoadOracleFixtureFS(bad, "d"); err == nil || !strings.Contains(err.Error(), fxA) {
		t.Errorf("malformed JSON: err = %v, want one naming the file", err)
	}
}

// TestOracleFixtureCoversRegistry's core: remove one catalogued card's
// file and exactly that card is reported, with a regen command that
// names only its ID.
func TestMissingOracleFileIsReported(t *testing.T) {
	have, err := LoadOracleFixture(OracleFixtureDir)
	if err != nil {
		t.Fatal(err)
	}
	specs := effects.All()
	if lines, _ := missingOracleFiles(specs, have); len(lines) != 0 {
		t.Fatalf("the checked-in fixture is already missing files: %v", lines)
	}
	victim, _ := BaseOracleID(specs[0].OracleID)
	delete(have, victim)
	lines, ids := missingOracleFiles(specs, have)
	if len(ids) != 1 || ids[0] != victim {
		t.Fatalf("missing = %v, want exactly [%s]", ids, victim)
	}
	if !strings.Contains(lines[0], OracleFixtureFile(victim)) {
		t.Errorf("the report %q does not name the file to add", lines[0])
	}
	cmd := oracleRegenCommand(ids)
	if !strings.Contains(cmd, "-update-oracle") || !strings.Contains(cmd, "-oracle-ids="+victim) {
		t.Errorf("regen command does not target the one card:\n%s", cmd)
	}
}

// TestOracleFixtureIsCurrent's core: a stale file for a card that left
// the catalog, a changed card, a missing card and a hand-reformatted
// file are all reported, and -oracle-ids narrows the report.
func TestOracleFixtureDiffCatchesStaleAndChangedFiles(t *testing.T) {
	a := OracleCard{Name: "A", Text: "a"}
	b := OracleCard{Name: "B", Text: "b"}
	c := OracleCard{Name: "C", Text: "c"}
	want := map[string]OracleCard{fxA: a, fxB: b, fxC: c}
	catalogued := map[string]bool{fxA: true, fxB: true, fxC: true}
	have := map[string]OracleCard{
		fxA: a,                            // current
		fxB: {Name: "B", Text: "erratum"}, // changed
		// fxC missing
		fxX: {Name: "X", Text: "x"}, // left the catalog
	}
	raw := map[string][]byte{}
	for id, card := range have {
		raw[id] = mustEncode(t, card)
	}

	diffs, ids := diffOracleFixture(want, have, raw, catalogued, nil)
	joined := strings.Join(diffs, "\n")
	for _, s := range []string{"changed   B (" + fxB, "missing   C (" + fxC, "not in catalog X (" + fxX} {
		if !strings.Contains(joined, s) {
			t.Errorf("diff does not report %q:\n%s", s, joined)
		}
	}
	if len(diffs) != 3 || !reflect.DeepEqual(ids, []string{fxB, fxC, fxX}) {
		t.Errorf("diffs = %v, ids = %v", diffs, ids)
	}

	// A hand-edit that keeps the data but not the generator's bytes.
	raw[fxA] = []byte("{\"name\": \"A\", \"text\": \"a\"}\n")
	if diffs, _ := diffOracleFixture(want, have, raw, catalogued, nil); !strings.Contains(strings.Join(diffs, "\n"), "formatting A") {
		t.Errorf("a reformatted file was not reported: %v", diffs)
	}
	raw[fxA] = mustEncode(t, a)

	// Scoped to A and B: C's absence and X's staleness are someone
	// else's cards.
	scope := map[string]bool{fxA: true, fxB: true}
	scopedWant := map[string]OracleCard{fxA: a, fxB: b}
	if _, ids := diffOracleFixture(scopedWant, have, raw, catalogued, scope); !reflect.DeepEqual(ids, []string{fxB}) {
		t.Errorf("scoped to A,B: ids = %v, want [%s]", ids, fxB)
	}
	// Naming the stale file does report it.
	if _, ids := diffOracleFixture(map[string]OracleCard{}, have, raw, catalogued, map[string]bool{fxX: true}); !reflect.DeepEqual(ids, []string{fxX}) {
		t.Errorf("scoped to X: ids = %v, want [%s]", ids, fxX)
	}
}

// The scoped writer touches only the named files: it rewrites a named
// stale card, removes a named uncatalogued file, and leaves every other
// file byte-identical. The full writer then does the rest — except a
// catalogued card the dump did not know, whose file it never deletes.
func TestScopedOracleRegenWritesOnlyNamedFiles(t *testing.T) {
	dir := t.TempDir()
	seed := map[string]string{
		fxA: "{\"name\":\"A\",\"text\":\"old\"}\n",
		fxB: "{\"name\":\"B\",\"text\":\"old\"}\n",
		fxC: "{\"name\":\"C\",\"text\":\"not in this dump\"}\n",
		fxX: "{\"name\":\"X\",\"text\":\"left the catalog\"}\n",
		fxY: "{\"name\":\"Y\",\"text\":\"left the catalog too\"}\n",
	}
	for id, body := range seed {
		if err := os.WriteFile(filepath.Join(dir, id+".json"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	read := func(id string) (string, bool) {
		b, err := os.ReadFile(filepath.Join(dir, id+".json"))
		return string(b), err == nil
	}
	want := map[string]OracleCard{fxA: {Name: "A", Text: "new"}, fxB: {Name: "B", Text: "new"}}
	catalogued := map[string]bool{fxA: true, fxB: true, fxC: true}

	wrote, removed, err := writeOracleFixture(dir, want, catalogued, map[string]bool{fxA: true, fxX: true})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wrote, []string{fxA}) || !reflect.DeepEqual(removed, []string{fxX}) {
		t.Fatalf("scoped: wrote %v removed %v, want [A] [X]", wrote, removed)
	}
	if got, _ := read(fxA); got != string(mustEncode(t, want[fxA])) {
		t.Errorf("A = %q, want the regenerated card", got)
	}
	if _, ok := read(fxX); ok {
		t.Error("X was named and is not catalogued, and its file survived")
	}
	for _, id := range []string{fxB, fxC, fxY} {
		if got, ok := read(id); !ok || got != seed[id] {
			t.Errorf("%s was not named and changed: %q", id, got)
		}
	}

	wrote, removed, err = writeOracleFixture(dir, want, catalogued, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wrote, []string{fxB}) || !reflect.DeepEqual(removed, []string{fxY}) {
		t.Fatalf("full: wrote %v removed %v, want [B] [Y]", wrote, removed)
	}
	if got, ok := read(fxC); !ok || got != seed[fxC] {
		t.Errorf("C is catalogued but missing from the dump; the writer touched its file (%q, exists=%v)", got, ok)
	}
	// A second full run is a no-op.
	if wrote, removed, _ := writeOracleFixture(dir, want, catalogued, nil); len(wrote)+len(removed) != 0 {
		t.Errorf("second run wrote %v removed %v", wrote, removed)
	}
}

func TestParseOracleIDs(t *testing.T) {
	got, err := parseOracleIDs(" " + strings.ToUpper(fxA) + "#1, " + fxB + ",")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, map[string]bool{fxA: true, fxB: true}) {
		t.Errorf("parseOracleIDs = %v", got)
	}
	if got, err := parseOracleIDs(""); got != nil || err != nil {
		t.Errorf("empty flag = %v, %v; want nil (every card)", got, err)
	}
	if _, err := parseOracleIDs("Lightning Bolt"); err == nil {
		t.Error("a card name was accepted as an oracle ID")
	}
}
