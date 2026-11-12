package game

import (
	"testing"

	"github.com/google/uuid"
)

// attacked_battle_tally_test.go — #5649: the turn's attack records say
// whether each attack was declared at a battle, so "as long as it
// attacked a battle this turn" (War Historian) can be asked after the
// battle has gone.

func attackedBattle(g *Game, id uuid.UUID) (b bool) {
	g.WithWriteLock(func() { b = g.AttackedBattleThisTurn(id) })
	return
}

// A creature declared attacking a battle attacked one, and still did
// once the battle has left the battlefield; one declared at a player or
// a planeswalker did not.
func TestAttackedBattleThisTurnReadsTheDeclaration(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	me := g.Seats[0].ID
	siege := pushBattleForTest(g, me, g.Seats[1].ID, "A Siege", 5)
	walker := pushPlaneswalkerForTest(g, g.Seats[2].ID, "Walker", 4)
	atBattle := pushReadyAttackerForTest(g, me, "At Battle", 2)
	atPlayer := pushReadyAttackerForTest(g, me, "At Player", 2)
	atWalker := pushReadyAttackerForTest(g, me, "At Walker", 2)
	advanceTo(t, g, StepDeclareAttackers)

	if attackedBattle(g, atBattle) {
		t.Fatal("nothing has attacked yet")
	}
	for id, target := range map[uuid.UUID]uuid.UUID{atBattle: siege, atPlayer: g.Seats[1].ID, atWalker: walker} {
		if err := g.DeclareAttacker(id, target); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	if attackedBattle(g, atBattle) {
		t.Fatal("a staged declaration is not an attack until it is locked in (CR 508.1)")
	}
	lockInAttackDeclaration(t, g)

	if !attackedBattle(g, atBattle) {
		t.Fatal("the creature declared at the battle did not read as having attacked one")
	}
	if attackedBattle(g, atPlayer) || attackedBattle(g, atWalker) {
		t.Error("an attack on a player or a planeswalker read as an attack on a battle")
	}

	g.WithWriteLock(func() { g.Battlefield.Remove(siege) })
	if !attackedBattle(g, atBattle) {
		t.Error("the attack stopped counting once the battle left the battlefield")
	}
}

// CR 508.7a: a reselected creature is "still considered to have attacked
// the player or permanent chosen as it was declared as an attacker", in
// both directions.
func TestAttackedBattleThisTurnIgnoresAReselect(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	me := g.Seats[0].ID
	siege := pushBattleForTest(g, me, g.Seats[1].ID, "A Siege", 5)
	onto := pushReadyAttackerForTest(g, me, "Moved Onto", 2)
	off := pushReadyAttackerForTest(g, me, "Moved Off", 2)
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(onto, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(off, siege); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)

	if err := reselect(g, onto, siege); err != nil {
		t.Fatalf("reselect onto the battle: %v", err)
	}
	if err := reselect(g, off, g.Seats[1].ID); err != nil {
		t.Fatalf("reselect off the battle: %v", err)
	}
	if attackedBattle(g, onto) {
		t.Error("a creature reselected onto a battle read as having attacked one")
	}
	if !attackedBattle(g, off) {
		t.Error("a creature reselected off the battle it was declared at stopped having attacked one")
	}
}

// CR 400.7: a creature that left and came back is a new object with no
// attacks; and the next turn's tally starts empty.
func TestAttackedBattleThisTurnIsPerObjectAndPerTurn(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	me := g.Seats[0].ID
	siege := pushBattleForTest(g, me, g.Seats[1].ID, "A Siege", 5)
	a := pushReadyAttackerForTest(g, me, "Attacker", 2)
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(a, siege); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)
	if !attackedBattle(g, a) {
		t.Fatal("setup: the attack was not recorded")
	}

	g.WithWriteLock(func() { findCard(g, a).ObjectEpoch++ })
	if attackedBattle(g, a) {
		t.Error("a new object with the same id inherited the old one's attack")
	}
	g.WithWriteLock(func() {
		findCard(g, a).ObjectEpoch--
		g.resetTurnTallyLocked()
	})
	if attackedBattle(g, a) {
		t.Error("the attack carried into the next turn's tally")
	}
}

// The flag is carried state: it survives a snapshot round trip, with the
// battle already gone, and an undo clone.
func TestAttackedBattleRecordRoundTrips(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	me := g.Seats[0].ID
	siege := pushBattleForTest(g, me, g.Seats[1].ID, "A Siege", 5)
	a := pushReadyAttackerForTest(g, me, "Attacker", 2)
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(a, siege); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)
	g.WithWriteLock(func() { g.Battlefield.Remove(siege) })

	_, restored := roundTrip(t, g)
	if !attackedBattle(restored, a) {
		t.Error("the restored game forgot the creature attacked a battle")
	}
	var clone *Game
	g.WithWriteLock(func() { clone = g.cloneLocked() })
	if !attackedBattle(clone, a) {
		t.Error("the undo clone forgot the creature attacked a battle")
	}
}
