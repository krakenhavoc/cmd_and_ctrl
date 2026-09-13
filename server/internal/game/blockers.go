package game

import "github.com/google/uuid"

// blockers.go models the declare-blockers turn-based action (#328).
//
// Background. Blocking in this engine is an ordinary priority-window
// action: runStepEntryHooksLocked has no StepDeclareBlockers case, so
// entering the step just grants priority to the active seat and lets
// it rotate. DeclareBlocker is a mutation a defending seat MAY send
// while that window is open. Nothing in the step machinery
// distinguishes "the defender chose not to block" from "the
// defender's client passed priority for them" — and per CR 509.1 the
// declaration is a turn-based action, not a response, so a
// legal-response predicate never sees it either.
//
// That is exactly how #328 happened: a defending player with
// auto-pass engaged had the window passed on their behalf and took
// eight unblocked damage with an untapped creature on the table.
//
// The fix keeps the rules right — declining to block is legal, and
// an explicit pass is how you decline — while guaranteeing a human
// sees the window. SeatOwesBlockDecision is the signal: "this seat is
// under attack and has a legal block available." The wire projection
// carries it to the client, which refuses to auto-pass while it
// holds. The server stays permissive, so no existing client, test, or
// bot is broken by a new rejection.

// BlockerEligible reports whether card b could be declared as a
// blocker by `seat` right now, ignoring which attacker it would be
// pointed at: b is a creature that seat controls, it is untapped
// (CR 509.1a), and it is not already blocking something.
//
// Summoning sickness deliberately does NOT disqualify a blocker.
// CR 302.6 restricts attacking and {T} / {Q} abilities only — a
// creature that arrived this turn blocks perfectly well. This is not
// a hypothetical: in the #328 replay the defender's one creature was
// summoning sick, and treating that as "can't block" would have
// justified the very skip that lost them the life.
//
// nil is ineligible so callers can skip defensive nil checks.
func BlockerEligible(b *Card, seat uuid.UUID) bool {
	if b == nil {
		return false
	}
	if b.Controller != seat || !b.IsCreature() {
		return false
	}
	if b.Tapped {
		return false
	}
	return b.BlockingTarget == uuid.Nil
}

// SeatOwesBlockDecision reports whether the seat with the given
// player ID is currently facing a declare-blockers decision it has
// not been given the chance to make: the cursor is on
// declare_blockers, at least one creature is attacking that seat, and
// the seat controls at least one creature that could legally be
// declared as a blocker against at least one of those attackers
// (CR 509.1a / 509.1b, evasion included via CanBlock).
//
// Deliberately NOT consumed by a block already declared. A defender
// who has assigned one blocker may still want to assign a second, so
// the answer stays true while any eligible creature remains — "stop
// once, then resume auto-passing" would slam the window shut on the
// first block, which is the same bug wearing a different hat.
//
// Returns false outside the declare_blockers step, for an unknown or
// eliminated seat, and for a seat with no legal block — in which case
// auto-passing the window costs the player nothing and is the right
// behaviour.
//
// Menace is deliberately not folded in. CanBlock is per-pair, while
// menace is a block-COUNT rule the engine enforces at the step's
// close-out (BlockerCountValid), so a defender holding exactly one
// eligible creature against a lone menace attacker is reported as
// owing a decision they cannot actually act on. That errs toward
// stopping, which is the safe direction for this signal: a spurious
// stop costs a click, a spurious skip costs the game.
//
// Goes through ReadSnapshot rather than taking the read lock
// directly, because the layer refresh it needs is a WRITE (see
// RecomputeLayersIfStaleLocked) and ReadSnapshot owns the upgrade.
// Callers must not hold g.mu. Use seatOwesBlockDecisionLocked from
// inside an existing lock.
func (g *Game) SeatOwesBlockDecision(seat uuid.UUID) bool {
	var out bool
	g.ReadSnapshot(func() {
		out = g.seatOwesBlockDecisionLocked(seat)
	})
	return out
}

// seatOwesBlockDecisionLocked is SeatOwesBlockDecision without the
// lock. Layers must already be fresh — CanBlock reads the effective
// characteristic, so flying granted by an anthem this turn has to be
// visible. Caller must hold g.mu (read or write).
func (g *Game) seatOwesBlockDecisionLocked(seat uuid.UUID) bool {
	if g.State != StateActive || g.Turn.Step != StepDeclareBlockers {
		return false
	}
	if seat == uuid.Nil {
		return false
	}
	// Attackers pointed at this seat. Collected first so the blocker
	// scan below can stop at the first legal pairing.
	//
	// S27: "pointed at this seat" is no longer a bare id comparison.
	// An attack on a planeswalker names the PLANESWALKER, and its
	// controller is the one who may block (CR 509.1a); an attack on a
	// battle names the battle, and its PROTECTOR blocks. Comparing
	// AttackingTarget to the seat directly would have told a player
	// under a full planeswalker assault that they owed no block
	// decision — and #328's auto-pass guard reads exactly this
	// function, so it would have passed the window for them.
	var attackers []*Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].AttackingTarget == uuid.Nil {
			continue
		}
		if g.defendingPlayerForAttackLocked(g.Battlefield.Cards[i].AttackingTarget) == seat {
			attackers = append(attackers, &g.Battlefield.Cards[i])
		}
	}
	if len(attackers) == 0 {
		return false
	}
	for i := range g.Battlefield.Cards {
		b := &g.Battlefield.Cards[i]
		if !BlockerEligible(b, seat) {
			continue
		}
		for _, a := range attackers {
			if CanBlock(a, b) {
				return true
			}
		}
	}
	return false
}

// SeatsOwingBlockDecisionLocked returns the seat INDICES that owe a
// declare-blockers decision right now, in seat order. The wire
// projection ships indices rather than player IDs to match TurnView's
// existing ActiveSeat / PriorityHolder shape.
//
// Eliminated seats are skipped — they are not being attacked and have
// nothing to defend.
//
// Caller must hold g.mu (read or write) with fresh layers — this is
// the read surface protocol.ViewOfGame calls from inside its existing
// ReadSnapshot, which guarantees the freshness.
func (g *Game) SeatsOwingBlockDecisionLocked() []int {
	if g.Turn.Step != StepDeclareBlockers {
		return nil
	}
	var out []int
	for i, s := range g.Seats {
		if s == nil || s.Eliminated {
			continue
		}
		if g.seatOwesBlockDecisionLocked(s.ID) {
			out = append(out, i)
		}
	}
	return out
}
