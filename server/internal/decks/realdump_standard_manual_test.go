package decks

import (
	"os"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestRealDumpStandardDecksAreLegal is the Standard half of
// realdump_manual_test.go, gated on CMDCTRL_SCRYFALL_DUMP the same way:
//
//	CMDCTRL_SCRYFALL_DUMP=data/scryfall/default-cards.json \
//	  go test ./internal/decks/ -run RealDumpStandard -v
//
// It is the only check of the card pool: Load runs deck.Validate, which
// reads legalities.standard from the dump. Run it after touching a
// Standard list and after a rotation or a ban.
func TestRealDumpStandardDecksAreLegal(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatalf("Load: %v", err)
	}
	for _, d := range In(game.FormatStandard) {
		t.Run(d.ID, func(t *testing.T) {
			list, err := d.Load(idx)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if len(list.Commanders) != 0 {
				t.Errorf("resolved %d commanders, want none", len(list.Commanders))
			}
			if got := len(list.Mainboard); got != 60 {
				t.Errorf("resolved to %d cards, want 60", got)
			}
		})
	}
}
