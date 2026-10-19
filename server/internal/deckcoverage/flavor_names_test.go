package deckcoverage

import (
	"os"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// The reported saved Squall list uses the titles printed on these nine
// reskins, while Moxfield's API returns their canonical names.
func TestRealDumpSquallAlternateNames(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to test the reported deck against Scryfall")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatal(err)
	}
	pairs := map[string]string{
		"Aang, Ascendant Airbender":       "Serra Ascendant",
		"Garnet Til Alexandros 17th":      "Loran of the Third Path",
		"Ghal Maraz, the Great Shatterer": "Loxodon Warhammer",
		"Hero of Light":                   "Adeline, Resplendent Cathar",
		"Paths of the Dead":               "Cavern of Souls",
		"Plankton, Tiny Tyrant":           "Skrelv, Defector Mite",
		"Raph's Jitte":                    "Umezawa's Jitte",
		"Squall Leonhart":                 "Danitha Capashen, Paragon",
		"Totec's Spear":                   "Shadowspear",
	}
	var entries []deck.Entry
	for alias, canonical := range pairs {
		entries = append(entries, deck.Entry{Name: alias, Count: 1})
		got, ok := idx.FindByName(alias)
		want, found := idx.FindByName(canonical)
		if !ok || !found || got.Name != canonical || got.OracleID != want.OracleID {
			t.Errorf("%s did not resolve as %s", alias, canonical)
		}
	}
	list, err := deck.Resolve(idx, "Squall", entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Mainboard) != 9 {
		t.Fatalf("resolved %d of nine cards", len(list.Mainboard))
	}
	report, err := Build(idx, Deck{Name: "Squall", Entries: entries})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Unknown) != 0 || report.UnknownCopies != 0 || report.CopiesTotal() != 9 {
		t.Fatalf("coverage failed to resolve the nine reskins: %+v", report)
	}
}
