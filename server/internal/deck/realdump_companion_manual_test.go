package deck

import (
	"os"
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// realdump_companion_manual_test.go — ADR 0172 against the real
// Scryfall dump, gated on CMDCTRL_SCRYFALL_DUMP like the other
// realdump tests:
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  scripts/go-docker.sh go test ./internal/deck -run TestRealDumpCompanion
func TestRealDumpCompanion(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Jegantha's condition on the real records of the edge cases.
	for name, fails := range map[string]bool{
		"Kitchen Finks":     true,
		"Fire // Ice":       true,
		"Bonecrusher Giant": false,
		"Valki, God of Lies // Tibalt, Cosmic Impostor": false,
		"Delver of Secrets // Insectile Aberration":     false,
		"Command Tower":            false,
		"Counterspell":             true,
		"Jegantha, the Wellspring": false,
	} {
		c, ok := idx.FindByName(name)
		if !ok {
			t.Errorf("%s: not in the dump", name)
			continue
		}
		if got := repeatsAManaSymbol(c); got != fails {
			t.Errorf("%s (%s, %q): repeats = %v, want %v", name, c.Layout, companionManaCosts(c), got, fails)
		}
	}

	// The import fixtures, resolved against the real cards: each reads
	// Jegantha as the companion.
	for _, tc := range []struct {
		file  string
		parse func([]byte) ([]Entry, error)
	}{
		{"archidekt_3401_companion_category.json", func(raw []byte) ([]Entry, error) { _, e, err := ParseArchidekt(raw); return e, err }},
		{"archidekt_3401_companion_in_sideboard.json", func(raw []byte) ([]Entry, error) { _, e, err := ParseArchidekt(raw); return e, err }},
		{"moxfield_3401_sisay_companion_and_sideboard.json", func(raw []byte) ([]Entry, error) { _, e, err := ParseMoxfield(raw); return e, err }},
		{"moxfield_3401_export_lone_sideboard.txt", func(raw []byte) ([]Entry, error) { return ParseText(string(raw)) }},
	} {
		entries, err := tc.parse(readFixture(t, tc.file))
		if err != nil {
			t.Errorf("%s: %v", tc.file, err)
			continue
		}
		list, err := Resolve(idx, "", entries)
		if err != nil {
			t.Errorf("%s: Resolve: %v", tc.file, err)
			continue
		}
		if list.Companion == nil || list.Companion.Name != "Jegantha, the Wellspring" {
			t.Errorf("%s: companion %v", tc.file, list.Companion)
			continue
		}
		var codes []string
		if err := Validate(list); err != nil {
			codes = violationCodes(err)
		}
		t.Logf("%s: %d cards, violations %v", tc.file, len(list.Commanders)+len(list.Mainboard), codes)
		if slices.Contains(codes, CodeUnsupportedMechanic) || slices.Contains(codes, CodeNotACompanion) {
			t.Errorf("%s: the companion itself was refused: %v", tc.file, codes)
		}
	}
}
