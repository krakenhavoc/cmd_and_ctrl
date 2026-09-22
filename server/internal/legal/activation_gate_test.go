package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// activation_gate_test.go — the enumerator half of #1210, in the
// #499/#618 agreement style: every activation the enumerator OFFERS
// is one the engine accepts, and every activation the engine REFUSES
// is one the enumerator does not offer.
//
// That is the whole reason both read the same function. A bot offered
// a Cursed-Totem'd ability picks it, is refused, and picks it again.

const (
	oracleCursedTotem   = "6225a704-430a-4f56-ad87-0e8d87f285f5"
	oracleKrenkoMobBoss = "68418069-f615-40ef-ae0d-764192acae00"
	oracleSolRingGate   = "6ad8011d-3471-4369-9d68-b264cc027487"
)

// TestCursedTotemRemovesCreatureActivationsFromTheEnumeration is the
// gate's #544 half, asserted on BOTH ability kinds — Cursed Totem
// prints no mana exemption, so the mana move has to go too.
func TestCursedTotemRemovesCreatureActivationsFromTheEnumeration(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	krenko := battlefieldCard(g, active, game.Card{
		Name: "Krenko, Mob Boss", TypeLine: "Legendary Creature — Goblin Warrior",
		OracleID: oracleKrenkoMobBoss, Power: 3, Toughness: 3,
	})
	elves := battlefieldCard(g, active, game.Card{
		Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		OracleID: oracleLlanowarElves, Power: 1, Toughness: 1,
	})
	solRing := battlefieldCard(g, active, game.Card{
		Name: "Sol Ring", TypeLine: "Artifact", OracleID: oracleSolRingGate,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	// Without the Totem all three are moves — otherwise the absences
	// below would prove nothing.
	moves := legal.EnumerateFor(g, active.ID)
	if len(activationsOf(moves, krenko)) == 0 {
		t.Fatal("Krenko's ability should be a move with nothing restricting it")
	}
	if !hasManaMoveFrom(moves, elves) {
		t.Fatal("the Elves' mana ability should be a move with nothing restricting it")
	}
	if !hasManaMoveFrom(moves, solRing) {
		t.Fatal("Sol Ring should be a move with nothing restricting it")
	}

	battlefieldCard(g, active, game.Card{
		Name: "Cursed Totem", TypeLine: "Artifact", OracleID: oracleCursedTotem,
	})

	moves = legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, krenko); len(acts) != 0 {
		t.Errorf("under a Cursed Totem, Krenko is still offered: %v", labels(acts))
	}
	if hasManaMoveFrom(moves, elves) {
		t.Error("under a Cursed Totem, the Elves' MANA ability is still offered — the card prints no mana exemption")
	}
	if !hasManaMoveFrom(moves, solRing) {
		t.Error("under a Cursed Totem, Sol Ring is not offered — the card names creatures, not artifacts")
	}
	// #544 the other way: everything still on the list is accepted.
	dispatchAll(t, g, active.ID, activationsOf(moves, solRing))
}
