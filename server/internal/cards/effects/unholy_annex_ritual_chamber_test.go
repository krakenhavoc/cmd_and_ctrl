package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const unholyAnnexOracle = "bd388ad9-a47b-4b0b-b94a-8e4343cd3de5"

// TestUnholyAnnexDrawsThenLosesTwoWithoutADemon casts the left half and
// lets the end step come.
func TestUnholyAnnexDrawsThenLosesTwoWithoutADemon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	room := roomCardC(me.ID, unholyAnnexOracle, "Unholy Annex", "{2}{B}", "Ritual Chamber", "{3}{B}{B}", "B")
	castRoomC(t, g, me.ID, room, 0)
	settleOrdering(t, g, me.ID)
	hand, life := me.Hand.Size(), me.Life
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-2 {
		t.Errorf("hand %d→%d, life %d→%d; want +1 card and -2 life with no Demon", hand, me.Hand.Size(), life, me.Life)
	}
}

// TestRitualChamberMakesADemonAndTheAnnexThenDrains unlocks the
// Chamber (a 6/6 black flying Demon), and at the end step the Annex
// drains each opponent for 2.
func TestRitualChamberMakesADemonAndTheAnnexThenDrains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	room := roomCardC(me.ID, unholyAnnexOracle, "Unholy Annex", "{2}{B}", "Ritual Chamber", "{3}{B}{B}", "B")
	castRoomC(t, g, me.ID, room, 1)
	settleOrdering(t, g, me.ID)
	id := findBattlefieldByName(g, "Demon")
	if id == uuid.Nil {
		t.Fatal("unlocking Ritual Chamber made no Demon")
	}
	d, _ := battlefieldCardByID(g, id)
	if d.Power != 6 || d.Toughness != 6 || !d.HasColor("B") || !containsString(effectiveAbilities(t, g, id), "flying") {
		t.Errorf("the token is %d/%d %v, want a 6/6 black flyer", d.Power, d.Toughness, d.Colors)
	}
	unlockDoorC(t, g, me.ID, room.InstanceID, game.DoorLeft)
	settleOrdering(t, g, me.ID)
	hand, life := me.Hand.Size(), me.Life
	var opps []int
	for _, s := range g.Seats[1:] {
		opps = append(opps, s.Life)
	}
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life+2 {
		t.Errorf("hand %d→%d, life %d→%d; want +1 card and +2 life with a Demon", hand, me.Hand.Size(), life, me.Life)
	}
	for i, s := range g.Seats[1:] {
		if s.Life != opps[i]-2 {
			t.Errorf("opponent %d: %d → %d, want -2", i, opps[i], s.Life)
		}
	}
}
