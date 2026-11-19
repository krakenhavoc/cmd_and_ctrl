package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gilded_light_test.go — Gilded Light (Scourge, #5842): "You gain shroud
// until end of turn." CR 702.18a, so unlike Dawn's Truce's hexproof the
// caster's OWN targeted spell can't reach them either.

const gildedLightOracle = "f7e7b2aa-9fd2-41dd-b11e-3a233530dc83"

func TestGildedLightIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(gildedLightOracle)
	if !ok {
		t.Fatal("Gilded Light is not registered")
	}
	if spec.Name != "Gilded Light" {
		t.Errorf("registered as %q", spec.Name)
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("completeness %v, want Full", spec.Completeness)
	}
}

// TestGildedLightShroudRefusesTheCastersOwnSpell is the card, end to
// end: once it resolves, the caster's own Bolt at themselves is refused
// (the half hexproof would allow), so is an opponent's, and the caster
// can still aim at somebody else.
func TestGildedLightShroudRefusesTheCastersOwnSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	castCatalogSpell(t, g, "Gilded Light", "Instant", gildedLightOracle, nil)
	passPriorityAroundTable(t, g)

	if got := playerAbilities(g, me); !hasPlayerAbility(got, KeywordShroud) {
		t.Fatalf("Gilded Light's caster has %v, want shroud", got)
	}
	if got := playerAbilities(g, opp); len(got) != 0 {
		t.Errorf("the opponent has %v, want nothing", got)
	}
	if err := boltAtPlayerErr(t, g, me, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("the caster's own Bolt at themselves: got %v, want ErrIllegalTarget", err)
	}
	if err := boltAtPlayerErr(t, g, opp, me.ID); err != game.ErrIllegalTarget {
		t.Errorf("an opponent's Bolt at the caster: got %v, want ErrIllegalTarget", err)
	}
	if err := boltAtPlayerErr(t, g, me, opp.ID); err != nil {
		t.Errorf("the caster's Bolt at an unshrouded opponent: %v", err)
	}
}

// TestGildedLightShroudEndsAtEndOfTurn: CR 514.2, the next turn the
// caster is an ordinary target again.
func TestGildedLightShroudEndsAtEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	castCatalogSpell(t, g, "Gilded Light", "Instant", gildedLightOracle, nil)
	passPriorityAroundTable(t, g)
	if !hasPlayerAbility(playerAbilities(g, me), KeywordShroud) {
		t.Fatal("no shroud after Gilded Light resolved")
	}

	endTurn(t, g)
	if got := playerAbilities(g, me); hasPlayerAbility(got, KeywordShroud) {
		t.Errorf("the shroud survived into the next turn: %v", got)
	}
	next := g.Seats[g.Turn.ActiveSeat]
	if err := boltAtPlayerErr(t, g, next, me.ID); err != nil {
		t.Errorf("a Bolt at the former Gilded Light caster next turn: %v", err)
	}
}

// TestGildedLightCyclesForACard: Cycling {2} (CR 702.29a) bins the card
// and draws one, from the hand.
func TestGildedLightCyclesForACard(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: gildedLightOracle})
	if len(abilities) != 1 || !abilities[0].Cycling || abilities[0].Cost.Mana != "{2}" ||
		!abilities[0].Cost.DiscardSelf || !game.AbilityFunctionsFromZone(abilities[0], game.ZoneHand) {
		t.Fatalf("Gilded Light's abilities %+v, want one Cycling {2} hand ability", abilities)
	}

	g := newCatalogGame(t)
	id, me := cycleFromHand(t, g, "Gilded Light", "Instant", gildedLightOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled Gilded Light is not in the graveyard")
	}
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != before+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", before, got)
	}
	if hasPlayerAbility(playerAbilities(g, me), KeywordShroud) {
		t.Error("cycling Gilded Light granted shroud; only casting it does")
	}
}
