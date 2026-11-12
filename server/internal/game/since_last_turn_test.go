package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// since_last_turn_test.go — #5652: whether a player has been dealt combat
// damage since the last turn they took, kept across the turn boundaries
// that replace the per-turn tally.

// dealDamageTo deals damage the way the engine announces it, through the
// event the tally listens to.
func dealDamageTo(g *Game, source, victim uuid.UUID, amount int, combat bool) {
	emit(g, Event{Kind: EventDealDamage, Source: source, Target: victim, Amount: amount, Combat: combat})
}

func combatDamagedSince(g *Game, player uuid.UUID) (ok bool) {
	g.WithWriteLock(func() { ok = g.DealtCombatDamageSinceLastTurnForEffect(player) })
	return ok
}

func TestCombatDamageSinceLastTurnSurvivesOtherTurnsAndEndsWithYours(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})

	if combatDamagedSince(g, b.ID) {
		t.Fatal("setup: nobody has been dealt combat damage")
	}
	// Damage that is not combat damage is not asked about.
	dealDamageTo(g, bear, b.ID, 3, false)
	if combatDamagedSince(g, b.ID) {
		t.Error("noncombat damage is not combat damage")
	}
	dealDamageTo(g, bear, b.ID, 2, true)
	if !combatDamagedSince(g, b.ID) {
		t.Fatal("seat 1 was dealt combat damage")
	}
	if combatDamagedSince(g, a.ID) || combatDamagedSince(g, g.Seats[2].ID) {
		t.Error("only seat 1 was dealt combat damage")
	}

	// Seat 1's own turn begins: its upkeep still reads the damage dealt
	// during seat 0's turn, though the per-turn tally is gone.
	passTurn(t, g)
	if g.Turn.ActiveSeat != 1 {
		t.Fatalf("setup: seat %d is active, want seat 1", g.Turn.ActiveSeat)
	}
	if g.CombatDamageTakenThisTurn(b.ID) != 0 {
		t.Fatal("setup: the per-turn tally was not replaced")
	}
	if !combatDamagedSince(g, b.ID) {
		t.Error("seat 1 was dealt combat damage since its last turn")
	}

	// As seat 1's turn ends, "since your last turn" starts over, and
	// the other players' turns leave it alone.
	passTurn(t, g)
	if combatDamagedSince(g, b.ID) {
		t.Error("seat 1's turn ended; nothing has been dealt since")
	}
	passTurn(t, g)
	passTurn(t, g)
	passTurn(t, g)
	if g.Turn.ActiveSeat != 1 {
		t.Fatalf("setup: seat %d is active, want seat 1 again", g.Turn.ActiveSeat)
	}
	if combatDamagedSince(g, b.ID) {
		t.Error("a round with no combat damage")
	}
}

// Combat damage dealt to a player during their own turn was dealt during
// their last turn, not since it, once that turn has ended.
func TestCombatDamageDuringYourOwnTurnIsNotSinceIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a := g.Seats[0]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})
	dealDamageTo(g, bear, a.ID, 2, true)
	if !combatDamagedSince(g, a.ID) {
		t.Fatal("seat 0 was dealt combat damage")
	}
	passTurn(t, g)
	if combatDamagedSince(g, a.ID) {
		t.Error("seat 0's turn ended after the damage")
	}
}

// Combat damage to a permanent a player controls is not damage to them.
func TestCombatDamageToAPermanentIsNotToItsController(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})
	walker := pushPermanent(g, Card{Name: "Walker", TypeLine: "Planeswalker — Test", Owner: b.ID, Controller: b.ID})
	dealDamageTo(g, bear, walker, 2, true)
	if combatDamagedSince(g, b.ID) {
		t.Error("damage to seat 1's planeswalker is not damage to seat 1")
	}
}

func TestCombatDamageSinceLastTurnSurvivesARestoreAndAClone(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})
	dealDamageTo(g, bear, b.ID, 2, true)

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !combatDamagedSince(restored, b.ID) {
		t.Error("the restored table forgot seat 1 was dealt combat damage")
	}
	if combatDamagedSince(restored, a.ID) {
		t.Error("the restored table thinks seat 0 was dealt combat damage")
	}

	c := g.Clone()
	passTurn(t, g)
	passTurn(t, g)
	if combatDamagedSince(g, b.ID) {
		t.Fatal("setup: seat 1's turn ended")
	}
	if !c.Seats[1].CombatDamagedSinceLastTurn {
		t.Error("the clone shares the original's record")
	}
}

// "You haven't been dealt combat damage since your last turn" reads
// nothing while it holds, and is refused once it does not: combat damage
// is a turn-based action (CR 510.2), so no item on the stack changes it.
func TestNoCombatDamageSinceLastTurnConditionReadsNothingWhileItHolds(t *testing.T) {
	f := newCondFixture(t)
	a := f.mustAccess(t, cond(FootprintCondNoCombatDamageSinceLastTurn, 0), drawOne)
	if a.boardReads || a.monarchRead {
		t.Errorf("read the board or the monarch: %+v", a)
	}
	f.me.CombatDamagedSinceLastTurn = true
	if _, ok := f.access(t, cond(FootprintCondNoCombatDamageSinceLastTurn, 0), drawOne); ok {
		t.Error("after combat damage the condition is false, and the step is refused")
	}
}
