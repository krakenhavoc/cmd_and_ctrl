package effects

import (
	"testing"

	"github.com/google/uuid"
)

// The "then transform any number of Human Werewolves" half of Tovolar,
// Dire Overlord's upkeep trigger is a real choice: a Human Werewolf
// without daybound is offered and turns over when picked, and one left
// unpicked stays as it was.
func TestTovolarsUpkeepTransformsThePickedHumanWerewolves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	importAndCast(t, g, dnbWerewolf(dnbTovolar).row(), me)
	monSettle(t, g)
	old := func(name, back string) uuid.UUID {
		row := werewolfRow(uuid.NewString(), name, "Creature — Human Werewolf", "{2}{G}", "", back, "Creature — Werewolf", "",
			"2", "2", "4", "4", []string{"G"})
		row.Keywords = nil
		id := importToHand(row, me)
		c, _ := cardInZone(me.Hand, id)
		c.Controller = me.ID
		me.Hand.Remove(id)
		pushBattlefieldCardWithTimestamp(g, c)
		return id
	}
	picked := old("Picked Villager", "Picked Wolf")
	kept := old("Kept Villager", "Kept Wolf")
	b12Creature(g, me.ID, "Wolf A", "Creature — Wolf", 2, 2)
	b12Creature(g, me.ID, "Wolf B", "Creature — Wolf", 2, 2)
	dnbAfterASpellTurn(t, g)
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID, picked)
	monSettle(t, g)
	if c, _ := battlefieldCard(g, picked); c.ActiveFace != 1 {
		t.Errorf("the picked Human Werewolf is on face %d, want it transformed", c.ActiveFace)
	}
	if c, _ := battlefieldCard(g, kept); c.ActiveFace != 0 {
		t.Errorf("the unpicked Human Werewolf is on face %d, want it left alone", c.ActiveFace)
	}
}
