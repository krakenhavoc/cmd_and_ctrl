package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const increasingDevotionOracle = "7a5ff4d4-27b7-47d4-ba88-970c63c4e3fb"

func TestIncreasingDevotionCreatesFiveFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	castCatalogSpell(t, g, "Increasing Devotion", "Sorcery", increasingDevotionOracle, nil)
	passPriorityAroundTable(t, g)

	if got := b43TokensNamed(g, me.ID, "Human"); got != 5 {
		t.Fatalf("Human tokens = %d, want 5 from a hand cast", got)
	}
}

func TestIncreasingDevotionCreatesTenFromFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := seedGraveyardCard(t, g, "Increasing Devotion", "Sorcery", increasingDevotionOracle)

	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if got := b43TokensNamed(g, me.ID, "Human"); got != 10 {
		t.Fatalf("Human tokens = %d, want 10 from a graveyard cast", got)
	}
	if !g.Exile.Contains(id) {
		t.Errorf("the flashed-back sorcery should be exiled")
	}
}
