package testsso

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const spellweaverDuoOracle = "1fd4204f-f627-4919-9d48-79b172f4560e"

// TestSpellweaverDuoReturnsATappedCreatureWhenItsMayIsYes: the optional
// ETB, answered yes, bounces the tapped creature to its owner's hand.
func TestSpellweaverDuoReturnsATappedCreatureWhenItsMayIsYes(t *testing.T) {
	effectstest.AssistCardSpec(t, spellweaverDuoOracle, "Spellweaver Duo")

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tapped := effectstest.PushVanillaCreature(g, opp.ID, "Tapped", 2, 2)
	effectstest.AssistCardTapped(t, g, tapped)

	effectstest.CastCatalogSpell(t, g, "Spellweaver Duo", "Creature — Human Wizard", spellweaverDuoOracle, nil)
	pick := game.TargetRef{Kind: game.TargetCard, ID: tapped}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, true)

	if effectstest.AssistCardOnBattlefield(g, tapped) {
		t.Fatal("the tapped creature is still on the battlefield")
	}
	if !effectstest.AssistCardInZone(g, tapped, game.ZoneHand) || !opp.Hand.Contains(tapped) {
		t.Error("the tapped creature did not go to its owner's hand")
	}
}

// TestSpellweaverDuoDoesNothingWhenItsMayIsNo: its target is chosen as
// it goes on the stack; answered no as it resolves, the tapped creature
// stays put.
func TestSpellweaverDuoDoesNothingWhenItsMayIsNo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tapped := effectstest.PushVanillaCreature(g, opp.ID, "Tapped", 2, 2)
	effectstest.AssistCardTapped(t, g, tapped)

	effectstest.CastCatalogSpell(t, g, "Spellweaver Duo", "Creature — Human Wizard", spellweaverDuoOracle, nil)
	pick := game.TargetRef{Kind: game.TargetCard, ID: tapped}
	effectstest.AssistCardSettle(t, g, me.ID, &pick, false)

	if !effectstest.AssistCardOnBattlefield(g, tapped) {
		t.Error("the tapped creature left the battlefield after the may was declined")
	}
}
