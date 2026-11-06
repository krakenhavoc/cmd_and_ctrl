package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// haste_deck_test.go is #4981's arena deck: BattleDeck with haste grants
// and creatures with a {T} ability, each dealt by a catalog key the
// catalog knows, and the contestant that turns #4324's price off.

func TestHasteBattleDeck(t *testing.T) {
	deck := botarena.HasteBattleDeck([16]byte{})
	if len(deck) != len(botarena.BattleDeck([16]byte{})) {
		t.Errorf("%d cards, want BattleDeck's %d", len(deck), len(botarena.BattleDeck([16]byte{})))
	}
	n := map[string]int{}
	for _, c := range deck {
		n[c.Name]++
		if c.OracleID == "" {
			continue
		}
		if s, ok := effects.Lookup(c.OracleID); !ok || s.Name != c.Name {
			t.Errorf("%s is dealt by %s, which the catalog does not register under that name", c.Name, c.OracleID)
		}
	}
	want := map[string]int{
		"Expedite": 6, "Hanweir Battlements": 4, "Frostwielder": 4, "Goblin Sharpshooter": 4, "Fervent Paincaster": 4,
		"Mountain": 20, "Bear": 4, "Ogre": 4, "Drake": 4, "Wurm": 6, "Lightning Bolt": 6,
	}
	for name, k := range want {
		if n[name] != k {
			t.Errorf("%d %s, want %d", n[name], name, k)
		}
	}
	if !botarena.IsSynthetic(botarena.HasteDeckID) {
		t.Error("haste-battle needs no Scryfall dump")
	}
}

// TestNoHasteTapContestant: `heuristic-nohastetap` is an arena name that
// round-trips, is a known heuristic variant, and is not a lobby tier.
func TestNoHasteTapContestant(t *testing.T) {
	got, err := botarena.ParseContestant(botarena.NoHasteTapContestant)
	if err != nil {
		t.Fatalf("ParseContestant(%q): %v", botarena.NoHasteTapContestant, err)
	}
	if want := (botarena.SeatSpec{Tier: tiers.Heuristic, Variant: botarena.VariantNoHasteTap}); got != want {
		t.Errorf("parsed %+v, want %+v", got, want)
	}
	if got.Contestant() != "heuristic-nohastetap" {
		t.Errorf("round-trips as %q", got.Contestant())
	}
	if _, err := tiers.Parse(botarena.NoHasteTapContestant); err == nil {
		t.Error("tiers.Parse accepts heuristic-nohastetap; it would be offered in the lobby")
	}
	cfg := botarena.Config{
		Seats: []botarena.SeatSpec{got, {Tier: tiers.Heuristic, Deck: botarena.HasteDeckID}},
		Games: 1,
	}
	cfg.Seats[0].Deck = botarena.HasteDeckID
	if err := cfg.Validate(); err != nil {
		t.Errorf("a heuristic-nohastetap seat on haste-battle was refused: %v", err)
	}
}
