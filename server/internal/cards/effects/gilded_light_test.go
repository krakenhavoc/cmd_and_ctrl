package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gildedLightOracle = "f7e7b2aa-9fd2-41dd-b11e-3a233530dc83"

// TestGildedLightRefusesTheCastersOwnTargetedSpell is the card through
// the engine (#5842, CR 702.18a): once it resolves its caster has
// shroud, so an opponent's Bolt at them is refused AND so is their own,
// which is where shroud parts from hexproof (Leyline of Sanctity's test
// proves the other side). The shroud is gone when the next turn begins.
func TestGildedLightRefusesTheCastersOwnTargetedSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	castCatalogSpell(t, g, "Gilded Light", "Instant", gildedLightOracle, nil)
	passPriorityAroundTable(t, g)

	if got := playerAbilities(g, me); !hasPlayerAbility(got, KeywordShroud) {
		t.Fatalf("Gilded Light's caster has %v, want shroud", got)
	}
	if err := boltAtPlayerErr(t, g, me, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("your OWN Bolt at yourself under shroud: got %v, want ErrIllegalTarget", err)
	}
	if err := boltAtPlayerErr(t, g, opp, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at a shrouded player: got %v, want ErrIllegalTarget", err)
	}
	// The other seat is untouched.
	if err := boltAtPlayerErr(t, g, me, opp.ID); err != nil {
		t.Errorf("a Bolt at the seat without shroud: %v", err)
	}

	advanceOneTurnForTest(t, g)
	if got := playerAbilities(g, me); hasPlayerAbility(got, KeywordShroud) {
		t.Errorf("shroud until end of turn outlived the turn: %v", got)
	}
}
