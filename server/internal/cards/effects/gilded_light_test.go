package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gilded_light_test.go — #5842: Gilded Light gives its caster shroud
// (CR 702.18a) until end of turn. Proved through the engine's targeting
// gate, from both sides: the caster's own Bolt at themselves is refused,
// which hexproof would allow, and so is an opponent's.

const gildedLightOracle = "f7e7b2aa-9fd2-41dd-b11e-3a233530dc83"

func TestGildedLightRefusesYourOwnSpellAndAnOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	castCatalogSpell(t, g, "Gilded Light", "Instant", gildedLightOracle, nil)
	passPriorityAroundTable(t, g)

	if got := playerAbilities(g, me); !hasPlayerAbility(got, KeywordShroud) {
		t.Fatalf("after Gilded Light the caster has %v, want shroud", got)
	}
	if got := playerAbilities(g, opp); len(got) != 0 {
		t.Errorf("the opponent was given %v", got)
	}

	// The half hexproof does not have: the caster's own spell.
	if err := boltAtPlayerErr(t, g, me, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("the caster's own Bolt at themselves: got %v, want ErrIllegalTarget", err)
	}
	if err := boltAtPlayerErr(t, g, opp, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at the caster: got %v, want ErrIllegalTarget", err)
	}
	if opponentMayTarget(g, me.ID, me.ID) {
		t.Error("the targeting gate lets the caster choose themselves")
	}
	// Only the caster is shrouded.
	if err := boltAtPlayerErr(t, g, me, opp.ID); err != nil {
		t.Errorf("the caster's Bolt at an opponent: %v", err)
	}
}

// TestGildedLightEndsAtEndOfTurn: "until end of turn", so on the next
// player's turn the caster is an ordinary target again.
func TestGildedLightEndsAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	castCatalogSpell(t, g, "Gilded Light", "Instant", gildedLightOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasPlayerAbility(playerAbilities(g, me), KeywordShroud) {
		t.Fatal("Gilded Light gave no shroud")
	}

	advanceToUpkeepOf(t, g, 1)
	if got := playerAbilities(g, me); hasPlayerAbility(got, KeywordShroud) {
		t.Errorf("on the next turn the caster still has %v", got)
	}
	if !opponentMayTarget(g, opp.ID, me.ID) {
		t.Error("on the next turn an opponent still can't target the caster")
	}
}

// TestGildedLightCyclesForTwo: Cycling {2} discards it and draws a card.
func TestGildedLightCyclesForTwo(t *testing.T) {
	g := newCatalogGame(t)
	seedLibrary(g.Seats[g.Turn.ActiveSeat], "A", "B")
	ons1Cycle(t, g, "Gilded Light", "Instant", gildedLightOracle, "{C}{C}")
}
