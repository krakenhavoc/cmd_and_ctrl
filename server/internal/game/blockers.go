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
// (CR 509.1a / 509.1b, evasion included via CanBlockLocked).
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
// Menace is deliberately not folded in. CanBlockLocked is per-pair, while
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
// lock. Layers must already be fresh — CanBlockLocked reads the effective
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
			if g.CanBlockLocked(a, b) {
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

// --- the block declaration's lock-in (#830) -----------------------
//
// CR 509.1 makes declaring blockers ONE turn-based action, and
// CR 509.2a puts the triggers it produces on the stack when the
// declaration is complete — before anyone receives priority. This
// engine's DeclareBlocker is a per-pair verb the defender may send
// repeatedly while the declare-blockers window is open (see the file
// comment), and the sandbox's "Re-declare blocker" re-points a
// blocker from one attacker to another.
//
// So the verb only STAGES the pairing. Nothing is announced until
// the declaration is locked in, which is the first point at which
// play moves on inside the step:
//
//   - runStateChecksLocked, the engine's "a player would receive
//     priority" boundary (a trick cast in the step, a resolution), and
//   - the two places the cursor can leave the step — AdvanceStep and
//     PassPriority's wrap — which call runStateChecksLocked
//     themselves when a declaration is still pending, so the lock-in
//     always happens INSIDE declare_blockers.
//
// commitBlockDeclarationLocked is the one place block-declaration
// triggers are harvested from: it emits the events, and the ordinary
// event harvester (triggers.go) does the rest. Before #830 the
// per-click EventBlock was the trigger-bearing event, so a blocker
// pointed at Cyberman Patrol and then re-pointed elsewhere left the
// Patrol's afflict trigger standing on an unblocked attacker.

// blockDeclarationPendingLocked reports whether the staged block
// declaration differs from what has already been announced — i.e.
// whether a lock-in would emit anything. Cheap battlefield scan;
// callers use it to avoid running a state-check pass for nothing.
//
// Caller must hold g.mu.
func (g *Game) blockDeclarationPendingLocked() bool {
	if g.Battlefield == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.BlockingTarget == uuid.Nil {
			continue
		}
		if g.announcedBlocks[c.InstanceID] != c.BlockingTarget {
			return true
		}
	}
	return false
}

// commitBlockDeclarationLocked locks the staged block declaration in:
// one EventBlock per (blocker, attacker) pair that has not been
// announced yet, then one EventBecomesBlocked per attacker that is
// newly blocked (CR 506.4 — an attacker is blocked once, however many
// creatures block it).
//
// Idempotent, and cheap when there is nothing to do, so it can sit at
// the top of runStateChecksLocked. Everything it emits lands in one
// event batch (nothing here opens a new one), which is what makes a
// whole declaration one occurrence for OncePerBatch — the same
// property DeclareAttacker relies on for Adeline (#854, ADR 0049).
//
// Both passes walk the battlefield in order, so the events of one
// declaration are deterministic and a replay reproduces them.
//
// Caller must hold g.mu in write mode.
func (g *Game) commitBlockDeclarationLocked() {
	if !g.blockDeclarationPendingLocked() {
		return
	}
	// Pass 1: the per-pair blocks. "Whenever this creature blocks"
	// (CR 509.3a) and "becomes blocked by a creature" read these, and
	// so does the public game log.
	type pairing struct {
		blocker  uuid.UUID
		attacker uuid.UUID
		actor    uuid.UUID
	}
	var fresh []pairing
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.BlockingTarget == uuid.Nil || g.announcedBlocks[c.InstanceID] == c.BlockingTarget {
			continue
		}
		fresh = append(fresh, pairing{blocker: c.InstanceID, attacker: c.BlockingTarget, actor: c.Controller})
	}
	if g.announcedBlocks == nil {
		g.announcedBlocks = map[uuid.UUID]uuid.UUID{}
	}
	for _, p := range fresh {
		g.announcedBlocks[p.blocker] = p.attacker
	}
	for _, p := range fresh {
		g.EmitEvent(Event{
			Kind:   EventBlock,
			Actor:  p.actor,
			Source: p.blocker,
			CardID: p.blocker,
			Target: p.attacker,
		})
	}
	// Pass 2: "becomes blocked", once per attacker. An attacker that
	// was already announced as blocked earlier in this combat does
	// not become blocked a second time when another creature is added
	// to its block (CR 506.4, CR 509.1h).
	if g.announcedBecameBlocked == nil {
		g.announcedBecameBlocked = map[uuid.UUID]bool{}
	}
	var blocked []pairing
	for _, p := range fresh {
		if g.announcedBecameBlocked[p.attacker] {
			continue
		}
		g.announcedBecameBlocked[p.attacker] = true
		blocked = append(blocked, p)
	}
	for _, p := range blocked {
		g.EmitEvent(Event{
			Kind:   EventBecomesBlocked,
			Actor:  p.actor,
			Source: p.attacker,
			CardID: p.attacker,
			Target: p.attacker,
		})
	}
}

// clearBlockAnnouncementsLocked forgets this combat's announcements.
// Called from clearCombatLocked, alongside the BlockingTarget wipe
// they describe: next combat's identical pairing is a new declaration
// and announces again. Caller must hold g.mu.
func (g *Game) clearBlockAnnouncementsLocked() {
	g.announcedBlocks = nil
	g.announcedBecameBlocked = nil
}
