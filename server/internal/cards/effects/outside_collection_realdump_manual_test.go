package effects

import (
	"os"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// outside_collection_realdump_manual_test.go — ADR 0198 against the real
// Scryfall dump, gated on CMDCTRL_SCRYFALL_DUMP so CI never loads it:
//
//	CMDCTRL_SCRYFALL_DUMP=../data/scryfall/default-cards.json \
//	  go test ./internal/cards/effects/ -run RealDumpCollection -v
//
// It pins that the suggestion walk finds a card for every printed wish
// filter on the real database, the engine-automated cards first, in a
// time a resolving spell can afford.
func TestRealDumpCollectionSuggestsForEveryWish(t *testing.T) {
	path := os.Getenv("CMDCTRL_SCRYFALL_DUMP")
	if path == "" {
		t.Skip("set CMDCTRL_SCRYFALL_DUMP to run against the real dump")
	}
	idx := cards.NewIndex()
	if _, err := idx.Load(path); err != nil {
		t.Fatal(err)
	}
	prev := game.OutsideCollection
	game.OutsideCollection = deck.NewCollection(idx)
	t.Cleanup(func() { game.OutsideCollection = prev })

	filters := map[string]func(game.Card) bool{
		"a sorcery card (Burning Wish)":                    func(c game.Card) bool { return c.IsSorcery() },
		"an Eldrazi card (Coax from the Blind Eternities)": func(c game.Card) bool { return c.HasSubtype("Eldrazi") },
		"an artifact card (Karn, the Great Creator)":       func(c game.Card) bool { return c.IsArtifact() },
		"a multicolored card (Glittering Wish)":            func(c game.Card) bool { return len(c.Colors) > 1 },
		"any card (Mastermind's Acquisition)":              nil,
	}
	for name, match := range filters {
		start := time.Now()
		got := game.CollectionSuggestionsForEffect(match, 8)
		took := time.Since(start)
		t.Logf("%s: %v in %v", name, got, took)
		if len(got) == 0 {
			t.Errorf("%s: nothing suggested", name)
		}
		if took > 2*time.Second {
			t.Errorf("%s: the walk took %v", name, took)
		}
		if match != nil {
			for _, n := range got {
				c, ok := game.OutsideCollection.Named(n)
				if !ok || !match(c) {
					t.Errorf("%s: suggested %q, which does not match", name, n)
				}
			}
		}
		if len(got) > 0 && !game.IsCatalogCard(mustOracle(t, got[0])) {
			t.Errorf("%s: the first suggestion %q is not an automated card", name, got[0])
		}
	}
	start := time.Now()
	found := game.OutsideCollection.Search("wish", 40)
	t.Logf("search %q: %d rows in %v", "wish", len(found), time.Since(start))
}

func mustOracle(t *testing.T, name string) string {
	t.Helper()
	c, ok := game.OutsideCollection.Named(name)
	if !ok {
		t.Fatalf("%q not in the collection", name)
	}
	return c.OracleID
}
