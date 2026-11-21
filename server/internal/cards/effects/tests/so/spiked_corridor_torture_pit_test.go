package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const spikedCorridorOracle = "e49b902f-a556-4dab-8928-93caf0a3f609"

// TestSpikedCorridorMakesThreeDevilsThatPingWhenTheyDie casts the left
// half and kills a Devil: its own trigger deals 1 to a chosen target.
func TestSpikedCorridorMakesThreeDevilsThatPingWhenTheyDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	room := effectstest.RoomCardC(me.ID, spikedCorridorOracle, "Spiked Corridor", "{3}{R}", "Torture Pit", "{3}{R}", "R")
	effectstest.CastRoomC(t, g, me.ID, room, 0)
	effectstest.SettleOrdering(t, g, me.ID)
	if n := effectstest.B16CountNamed(g, "Devil"); n != 3 {
		t.Fatalf("%d Devils, want 3", n)
	}
	id := effectstest.FindBattlefieldByName(g, "Devil")
	d, _ := effectstest.BattlefieldCardByID(g, id)
	if d.Power != 1 || d.Toughness != 1 || !d.HasColor("R") {
		t.Errorf("the Devil is %d/%d %v, want a 1/1 red", d.Power, d.Toughness, d.Colors)
	}
	before := opp.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("a Devil dying: %d → %d, want -1", before, opp.Life)
	}
	if n := effectstest.B16CountNamed(g, "Devil"); n != 2 {
		t.Errorf("%d Devils left, want 2", n)
	}
}

// TestTorturePitAddsTwoToNoncombatDamageAtAnOpponent casts the right
// half: a red Bolt at an opponent deals 3 + 2, one at yourself 3, and
// combat damage is untouched. With only the Corridor unlocked the
// replacement does not exist.
func TestTorturePitAddsTwoToNoncombatDamageAtAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	room := effectstest.RoomCardC(me.ID, spikedCorridorOracle, "Spiked Corridor", "{3}{R}", "Torture Pit", "{3}{R}", "R")
	effectstest.CastRoomC(t, g, me.ID, room, 0)
	effectstest.SettleOrdering(t, g, me.ID)
	oppLife := opp.Life
	effectstest.B08CastBolt(t, g, opp.ID, true)
	if opp.Life != oppLife-3 {
		t.Fatalf("Torture Pit locked: Bolt took %d, want 3", oppLife-opp.Life)
	}

	effectstest.UnlockDoorC(t, g, me.ID, room.InstanceID, game.DoorRight)
	effectstest.SettleOrdering(t, g, me.ID)
	oppLife, myLife := opp.Life, me.Life
	effectstest.B08CastBolt(t, g, opp.ID, true)
	if opp.Life != oppLife-5 {
		t.Errorf("Bolt at an opponent: took %d, want 5", oppLife-opp.Life)
	}
	effectstest.B08CastBolt(t, g, me.ID, true)
	if me.Life != myLife-3 {
		t.Errorf("Bolt at yourself: took %d, want 3 (not an opponent)", myLife-me.Life)
	}
	goblin := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Creature — Goblin", Colors: []string{"R"},
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	oppLife = opp.Life
	effectstest.AttackWith(t, g, opp.ID, goblin)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 {
		t.Errorf("combat damage: took %d, want 1 (noncombat only)", oppLife-opp.Life)
	}
}
