package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// last_turn_attacks_test.go — ADR 0108 §6 (#1882): what a player's
// creatures attacked during that player's last turn, kept across the
// turn boundary that flushes TurnTally.Attacks.

// declareAttack records one attack the way the engine does, through the
// event the tally listens to.
func declareAttack(g *Game, attacker, defender uuid.UUID) {
	emit(g, Event{Kind: EventAttack, Actor: g.Seats[g.Turn.ActiveSeat].ID, CardID: attacker, Target: defender})
}

func lastTurnAsked(g *Game, c uuid.UUID) (ok bool) {
	g.WithWriteLock(func() {
		card := g.findCardByIDLocked(c)
		ok = g.AttackedDuringControllersLastTurn(card)
	})
	return ok
}

func attackedYouLast(g *Game, player, you uuid.UUID) (ok bool) {
	g.WithWriteLock(func() { ok = g.AttackedYouDuringTheirLastTurn(player, you) })
	return ok
}

func TestLastTurnAttacksSurviveTheTurnBoundary(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})

	if lastTurnAsked(g, bear) {
		t.Fatal("setup: nothing has attacked yet")
	}
	declareAttack(g, bear, b.ID)
	// This turn is not yet anybody's LAST turn.
	if lastTurnAsked(g, bear) {
		t.Fatal("an attack made this turn is not an attack during your LAST turn")
	}
	passTurn(t, g) // seat 1's turn begins; the tally is replaced
	if len(g.TurnTally.Attacks) != 0 {
		t.Fatalf("setup: the tally still holds %d attacks", len(g.TurnTally.Attacks))
	}
	if !lastTurnAsked(g, bear) {
		t.Error("the Bear attacked during its controller's last turn")
	}
	if !attackedYouLast(g, a.ID, b.ID) {
		t.Error("seat 0 attacked seat 1 during its last turn")
	}
	if attackedYouLast(g, a.ID, g.Seats[2].ID) {
		t.Error("seat 0 did not attack seat 2")
	}
	if attackedYouLast(g, b.ID, a.ID) {
		t.Error("seat 1 has not taken a turn yet")
	}

	// Other players' turns do not touch seat 0's record ("and not your
	// opponent's", the Giant Turtle ruling): it stays through seats 1-3.
	passTurn(t, g)
	passTurn(t, g)
	passTurn(t, g)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: seat %d is active, want seat 0 again", g.Turn.ActiveSeat)
	}
	if !lastTurnAsked(g, bear) {
		t.Error("three other turns later, seat 0's last turn is still the one it attacked in")
	}

	// A turn with no attack is the new last turn: the record is
	// replaced by the empty one.
	passTurn(t, g)
	if lastTurnAsked(g, bear) {
		t.Error("seat 0's last turn had no attack; last turn's attack is forgotten")
	}
	if attackedYouLast(g, a.ID, b.ID) {
		t.Error("the attack on seat 1 was two turns ago")
	}
}

// An attack on a planeswalker is not an attack on its controller, and
// a creature put onto the battlefield attacking was never declared
// (CR 508.4, which is why the record is made from declarations).
func TestAttackedYouMeansThePlayerNotTheirPlaneswalker(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})
	walker := pushPermanent(g, Card{Name: "Walker", TypeLine: "Planeswalker — Test", Owner: b.ID, Controller: b.ID})
	declareAttack(g, bear, walker)
	passTurn(t, g)
	if attackedYouLast(g, a.ID, b.ID) {
		t.Error("attacking Seat 1's planeswalker is not attacking Seat 1")
	}
	if !lastTurnAsked(g, bear) {
		t.Error("the Bear still attacked")
	}
}

// CR 400.7: a creature that left the battlefield and came back has no
// history, and the new object is asked of its own controller.
func TestLastTurnAttackIsKeyedByObjectEpoch(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})
	declareAttack(g, bear, b.ID)
	passTurn(t, g)
	g.WithWriteLock(func() { g.findCardByIDLocked(bear).ObjectEpoch++ })
	if lastTurnAsked(g, bear) {
		t.Error("a creature that changed zones is a new object with no attacks of its own")
	}
}

// "Your last turn" is the last turn YOU took: an extra turn is a turn
// (CR 500.7), so the turn after it reads the extra turn, and the
// record of the normal turn before it is replaced.
func TestAnExtraTurnReplacesTheRecord(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})
	takeExtraTurns(t, g, 0, 1)
	declareAttack(g, bear, b.ID)
	passTurn(t, g) // into the extra turn, seat 0 again
	if !g.Turn.Extra || g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: turn = %+v, want seat 0's extra turn", g.Turn)
	}
	if !lastTurnAsked(g, bear) {
		t.Fatal("the turn before the extra turn is its last turn")
	}
	passTurn(t, g) // the extra turn ends without an attack
	if lastTurnAsked(g, bear) {
		t.Error("the extra turn is seat 0's last turn now, and nothing attacked in it")
	}
}

// A creature's controller is read now: a creature that attacked for
// its old controller does not count for the new one, whose own last
// turn it did not attack in.
func TestLastTurnAttackReadsTheCurrentControllersTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})
	declareAttack(g, bear, b.ID)
	passTurn(t, g)
	g.WithWriteLock(func() { g.findCardByIDLocked(bear).Controller = g.Seats[2].ID })
	if lastTurnAsked(g, bear) {
		t.Error("seat 2 did not attack with it during seat 2's last turn")
	}
}

// The record is restore state: it is not derivable from the board.
func TestLastTurnAttacksRoundTripThroughASnapshot(t *testing.T) {
	g := newRestorableGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})
	declareAttack(g, bear, b.ID)
	passTurn(t, g)

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
	if got := restored.Seats[0].LastTurnAttacks; len(got) != 1 || got[0].Attacker != bear || got[0].Defender != b.ID {
		t.Errorf("restored LastTurnAttacks = %+v, want the Bear's attack on seat 1", got)
	}
	if !lastTurnAsked(restored, bear) {
		t.Error("the restored game forgot the Bear attacked during its controller's last turn")
	}
}

// An undo restore point holds its own copy: the live record is
// replaced as the next turn ends, never edited in place.
func TestCloneKeepsItsOwnLastTurnAttacks(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	bear := pushPermanent(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID})
	declareAttack(g, bear, b.ID)
	passTurn(t, g)
	c := g.Clone()
	if len(c.Seats[0].LastTurnAttacks) != 1 {
		t.Fatalf("clone LastTurnAttacks = %+v", c.Seats[0].LastTurnAttacks)
	}
	g.Seats[0].LastTurnAttacks[0].Attacker = uuid.New()
	if c.Seats[0].LastTurnAttacks[0].Attacker != bear {
		t.Error("the clone shares the live record's backing array")
	}
}
