package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const spellweaverDuoOracle = "1fd4204f-f627-4919-9d48-79b172f4560e"

// TestSpellweaverDuoReturnsATappedCreatureWhenItsMayIsYes: the optional
// ETB, answered yes, bounces the tapped creature to its owner's hand.
func TestSpellweaverDuoReturnsATappedCreatureWhenItsMayIsYes(t *testing.T) {
	assistCardSpec(t, spellweaverDuoOracle, "Spellweaver Duo")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tapped := pushVanillaCreature(g, opp.ID, "Tapped", 2, 2)
	assistCardTapped(t, g, tapped)

	castCatalogSpell(t, g, "Spellweaver Duo", "Creature — Human Wizard", spellweaverDuoOracle, nil)
	pick := game.TargetRef{Kind: game.TargetCard, ID: tapped}
	assistCardSettle(t, g, me.ID, &pick, true)

	if assistCardOnBattlefield(g, tapped) {
		t.Fatal("the tapped creature is still on the battlefield")
	}
	if !assistCardInZone(g, tapped, game.ZoneHand) || !opp.Hand.Contains(tapped) {
		t.Error("the tapped creature did not go to its owner's hand")
	}
}

// TestSpellweaverDuoDoesNothingWhenItsMayIsNo: answered no, the trigger
// makes no pick and the tapped creature stays put.
func TestSpellweaverDuoDoesNothingWhenItsMayIsNo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tapped := pushVanillaCreature(g, opp.ID, "Tapped", 2, 2)
	assistCardTapped(t, g, tapped)

	castCatalogSpell(t, g, "Spellweaver Duo", "Creature — Human Wizard", spellweaverDuoOracle, nil)
	assistCardSettle(t, g, me.ID, nil, false)

	if !assistCardOnBattlefield(g, tapped) {
		t.Error("the tapped creature left the battlefield after the may was declined")
	}
}
