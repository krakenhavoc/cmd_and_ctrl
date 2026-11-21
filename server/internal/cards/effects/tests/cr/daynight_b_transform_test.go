package testscr

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// The "then transform any number of Human Werewolves" half of Tovolar,
// Dire Overlord's upkeep trigger is a real choice: a Human Werewolf
// without daybound is offered and turns over when picked, and one left
// unpicked stays as it was.
func TestTovolarsUpkeepTransformsThePickedHumanWerewolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ImportAndCast(t, g, effectstest.DnbWerewolf(effectstest.DnbTovolar).Row(), me)
	effectstest.MonSettle(t, g)
	old := func(name, back string) uuid.UUID {
		row := effectstest.WerewolfRow(uuid.NewString(), name, "Creature — Human Werewolf", "{2}{G}", "", back, "Creature — Werewolf", "",
			"2", "2", "4", "4", []string{"G"})
		row.Keywords = nil
		id := effectstest.ImportToHand(row, me)
		c, _ := effectstest.CardInZone(me.Hand, id)
		c.Controller = me.ID
		me.Hand.Remove(id)
		effectstest.PushBattlefieldCardWithTimestamp(g, c)
		return id
	}
	picked := old("Picked Villager", "Picked Wolf")
	kept := old("Kept Villager", "Kept Wolf")
	effectstest.B12Creature(g, me.ID, "Wolf A", "Creature — Wolf", 2, 2)
	effectstest.B12Creature(g, me.ID, "Wolf B", "Creature — Wolf", 2, 2)
	dnbAfterASpellTurn(t, g)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, picked)
	effectstest.MonSettle(t, g)
	if c, _ := effectstest.BattlefieldCard(g, picked); c.ActiveFace != 1 {
		t.Errorf("the picked Human Werewolf is on face %d, want it transformed", c.ActiveFace)
	}
	if c, _ := effectstest.BattlefieldCard(g, kept); c.ActiveFace != 0 {
		t.Errorf("the unpicked Human Werewolf is on face %d, want it left alone", c.ActiveFace)
	}
}
