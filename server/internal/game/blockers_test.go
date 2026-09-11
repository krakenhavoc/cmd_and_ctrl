package game

import (
	"testing"

	"github.com/google/uuid"
)

// blockers_test.go covers #328 — the declare-blockers turn-based
// action. The engine models blocking as an ordinary priority-window
// action, so nothing in the step machinery distinguished "the
// defender chose not to block" from "the defender's client passed
// for them". SeatOwesBlockDecision is the signal that closes that
// gap: it says a seat is under attack with a legal block available,
// which the client uses to refuse to auto-pass the window.

// pushCreatureFor drops a plain creature on the battlefield under
// owner's control and returns its instance ID.
func pushCreatureFor(t *testing.T, g *Game, owner *Player, power, toughness int) uuid.UUID {
	t.Helper()
	return pushKeywordCreature(t, g, owner, power, toughness)
}

// TestSeatOwesBlockDecisionUnderAttack is the #328 reproduction.
// Seat 1 attacks seat 0; seat 0 has one untapped creature. Seat 0
// owes a block decision and the client must not auto-pass it.
func TestSeatOwesBlockDecisionUnderAttack(t *testing.T) {
	g := newActiveGame(t)
	defender := g.Seats[0]
	attackerSeat := g.Seats[1]

	// Hand the turn to seat 1 so it is the active (attacking) seat.
	g.Turn.ActiveSeat = 1
	attackerID := pushKeywordCreature(t, g, attackerSeat, 4, 4)
	blockerID := pushCreatureFor(t, g, defender, 2, 2)

	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attackerID, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	if !g.SeatOwesBlockDecision(defender.ID) {
		t.Fatal("defender under attack with an untapped creature must owe a block decision")
	}
	// The attacking seat is not being attacked — it owes nothing.
	if g.SeatOwesBlockDecision(attackerSeat.ID) {
		t.Error("attacking seat must not owe a block decision")
	}
	_ = blockerID
}

// TestSeatOwesBlockDecisionSummoningSickBlocker is the exact board
// from the #328 replay: the defender's only creature came down this
// turn. CR 302.6 restricts attacking and {T} abilities — NOT
// blocking — so a summoning-sick creature is still a legal blocker
// and the window must still stop.
func TestSeatOwesBlockDecisionSummoningSickBlocker(t *testing.T) {
	g := newActiveGame(t)
	defender := g.Seats[0]
	attackerSeat := g.Seats[1]
	g.Turn.ActiveSeat = 1

	attackerID := pushKeywordCreature(t, g, attackerSeat, 4, 4)
	blockerID := pushCreatureFor(t, g, defender, 2, 2)
	// Mark the defender's creature as having arrived this turn.
	findCard(g, blockerID).SummonedThisTurn = true
	if !HasSummoningSickness(findCard(g, blockerID)) {
		t.Fatal("test setup: blocker should read as summoning sick")
	}

	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attackerID, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	if !g.SeatOwesBlockDecision(defender.ID) {
		t.Fatal("summoning-sick creature is a legal blocker (CR 302.6) — window must stop")
	}
}

// TestSeatOwesBlockDecisionNoLegalBlock covers the cases where the
// window really is empty, so auto-pass is free to skip it: no
// attackers, a tapped-only board, and an evasion mismatch.
func TestSeatOwesBlockDecisionNoLegalBlock(t *testing.T) {
	t.Run("no attackers", func(t *testing.T) {
		g := newActiveGame(t)
		defender := g.Seats[0]
		g.Turn.ActiveSeat = 1
		pushCreatureFor(t, g, defender, 2, 2)
		advanceIntoStep(t, g, StepDeclareBlockers)
		if g.SeatOwesBlockDecision(defender.ID) {
			t.Error("nothing is attacking — no decision is owed")
		}
	})

	t.Run("only blocker is tapped", func(t *testing.T) {
		g := newActiveGame(t)
		defender := g.Seats[0]
		attackerSeat := g.Seats[1]
		g.Turn.ActiveSeat = 1
		attackerID := pushKeywordCreature(t, g, attackerSeat, 4, 4)
		blockerID := pushCreatureFor(t, g, defender, 2, 2)

		advanceIntoStep(t, g, StepDeclareAttackers)
		if err := g.DeclareAttacker(attackerID, defender.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		findCard(g, blockerID).Tapped = true
		advanceIntoStep(t, g, StepDeclareBlockers)
		if g.SeatOwesBlockDecision(defender.ID) {
			t.Error("a tapped creature can't block (CR 509.1a) — no decision is owed")
		}
	})

	t.Run("flying attacker, ground-only defender", func(t *testing.T) {
		g := newActiveGame(t)
		defender := g.Seats[0]
		attackerSeat := g.Seats[1]
		g.Turn.ActiveSeat = 1
		attackerID := pushKeywordCreature(t, g, attackerSeat, 4, 4, "flying")
		pushCreatureFor(t, g, defender, 2, 2)

		advanceIntoStep(t, g, StepDeclareAttackers)
		if err := g.DeclareAttacker(attackerID, defender.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceIntoStep(t, g, StepDeclareBlockers)
		if g.SeatOwesBlockDecision(defender.ID) {
			t.Error("no flying / reach blocker — no legal block, no decision owed")
		}
	})

	t.Run("reach blocker answers a flier", func(t *testing.T) {
		g := newActiveGame(t)
		defender := g.Seats[0]
		attackerSeat := g.Seats[1]
		g.Turn.ActiveSeat = 1
		attackerID := pushKeywordCreature(t, g, attackerSeat, 4, 4, "flying")
		pushKeywordCreature(t, g, defender, 2, 2, "reach")

		advanceIntoStep(t, g, StepDeclareAttackers)
		if err := g.DeclareAttacker(attackerID, defender.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceIntoStep(t, g, StepDeclareBlockers)
		if !g.SeatOwesBlockDecision(defender.ID) {
			t.Error("reach blocks flying — the window must stop")
		}
	})
}

// TestSeatOwesBlockDecisionOutsideDeclareBlockers pins the step
// gate: the signal is only meaningful during declare_blockers, so
// every other step reports false and the client's auto-pass is
// untouched there.
func TestSeatOwesBlockDecisionOutsideDeclareBlockers(t *testing.T) {
	g := newActiveGame(t)
	defender := g.Seats[0]
	attackerSeat := g.Seats[1]
	g.Turn.ActiveSeat = 1
	attackerID := pushKeywordCreature(t, g, attackerSeat, 4, 4)
	pushCreatureFor(t, g, defender, 2, 2)

	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attackerID, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if g.SeatOwesBlockDecision(defender.ID) {
		t.Error("declare_attackers is not the blocking window")
	}
	advanceIntoStep(t, g, StepCombatDamage)
	if g.SeatOwesBlockDecision(defender.ID) {
		t.Error("combat_damage is not the blocking window")
	}
}

// TestSeatOwesBlockDecisionAfterDeclaringOneBlocker is the trap that
// makes "stop once" wrong. A defender who has declared one blocker
// may still want to declare a second; the window must keep holding
// while any eligible creature remains, otherwise auto-pass slams it
// shut on the first block.
func TestSeatOwesBlockDecisionAfterDeclaringOneBlocker(t *testing.T) {
	g := newActiveGame(t)
	defender := g.Seats[0]
	attackerSeat := g.Seats[1]
	g.Turn.ActiveSeat = 1
	attackerID := pushKeywordCreature(t, g, attackerSeat, 6, 6)
	first := pushCreatureFor(t, g, defender, 2, 2)
	pushCreatureFor(t, g, defender, 2, 2)

	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attackerID, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(first, attackerID); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if !g.SeatOwesBlockDecision(defender.ID) {
		t.Fatal("a second eligible blocker remains — the window must keep holding")
	}
}

// TestBlockerEligible pins the per-card predicate the enumerator and
// the wire projection share.
func TestBlockerEligible(t *testing.T) {
	seat := uuid.New()
	other := uuid.New()

	creature := func() *Card {
		c := NewCard("Blocker", seat)
		c.TypeLine = "Creature — Test"
		c.Controller = seat
		return &c
	}

	if !BlockerEligible(creature(), seat) {
		t.Error("untapped creature you control is an eligible blocker")
	}
	tapped := creature()
	tapped.Tapped = true
	if BlockerEligible(tapped, seat) {
		t.Error("tapped creature is not an eligible blocker (CR 509.1a)")
	}
	sick := creature()
	sick.SummonedThisTurn = true
	if !BlockerEligible(sick, seat) {
		t.Error("summoning sickness does not stop a creature blocking (CR 302.6)")
	}
	blocking := creature()
	blocking.BlockingTarget = uuid.New()
	if BlockerEligible(blocking, seat) {
		t.Error("a creature already blocking is not eligible again")
	}
	notMine := creature()
	notMine.Controller = other
	if BlockerEligible(notMine, seat) {
		t.Error("a creature you don't control is not your blocker")
	}
	land := creature()
	land.TypeLine = "Land"
	if BlockerEligible(land, seat) {
		t.Error("a non-creature is not a blocker")
	}
	if BlockerEligible(nil, seat) {
		t.Error("nil card must be ineligible")
	}
}
