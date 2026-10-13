package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const vampireCharmseekerOracle = "993bc179-4e24-4161-866f-5e146cbdd2f5"

// TestVampireCharmseekerReturnsAnInstantFromAnOpponentsGraveyard: the
// pick reads every graveyard, and the card goes to its owner's hand.
func TestVampireCharmseekerReturnsAnInstantFromAnOpponentsGraveyard(t *testing.T) {
	assistCardSpec(t, vampireCharmseekerOracle, "Vampire Charmseeker")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opt := pushGraveyardTyped(opp, "Opt", "Instant")

	castCatalogSpell(t, g, "Vampire Charmseeker", "Creature — Vampire Wizard", vampireCharmseekerOracle, nil)
	pick := game.TargetRef{Kind: game.TargetCard, ID: opt}
	assistCardSettle(t, g, me.ID, &pick, true)

	if !opp.Hand.Contains(opt) {
		t.Error("the instant did not return to its owner's hand")
	}
	if opp.Graveyard.Contains(opt) {
		t.Error("the instant is still in the graveyard")
	}
}

// TestVampireCharmseekerReturnsACreatureCardFromTheCastersGraveyard:
// the creature choice takes a creature card from the caster's own
// graveyard, which goes to the caster's hand.
func TestVampireCharmseekerReturnsACreatureCardFromTheCastersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushGraveyardTyped(me, "Bear", "Creature — Bear")

	castCatalogSpell(t, g, "Vampire Charmseeker", "Creature — Vampire Wizard", vampireCharmseekerOracle, nil)
	pick := game.TargetRef{Kind: game.TargetCard, ID: bear}
	assistCardSettle(t, g, me.ID, &pick, true)

	if !me.Hand.Contains(bear) {
		t.Error("the creature card did not return to its owner's hand")
	}
}
