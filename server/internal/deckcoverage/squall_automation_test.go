package deckcoverage

import (
	"os"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

func TestRealDumpSquallDeckFullyAutomated(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to check the reported Squall deck")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("testdata/squall_4126.txt")
	if err != nil {
		t.Fatal(err)
	}
	entries, err := deck.ParseText(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := deck.Resolve(idx, "Squall", entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := deck.Validate(resolved); err != nil {
		t.Fatalf("deck validation: %v", err)
	}
	report, err := Build(idx, Deck{Name: "Squall", Entries: entries})
	if err != nil {
		t.Fatal(err)
	}
	if report.CopiesTotal() != 100 || report.UnknownCopies != 0 || len(report.Unknown) != 0 {
		t.Fatalf("deck size or unresolved names: %+v", report)
	}
	for _, card := range report.Cards {
		if card.Bucket != Automated && card.Bucket != NoEffect {
			t.Errorf("%s is still %s: %v", card.Name, card.Bucket, card.Caveats)
		}
	}
	automated, total := report.AsPrintedCopies()
	if automated != 100 || total != 100 {
		t.Errorf("only %d/%d cards play as printed", automated, total)
	}
}
