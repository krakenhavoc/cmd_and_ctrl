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
// Block COUNTS are folded in (#750, ADR 0045 addendum Decision 14).
// The signal asks the same option generator the legal-move enumerator
// asks — BlockOptionsLocked — so "this seat has a legal block" means
// the same thing to both, and to the verb that would accept it. A
// defender whose only untapped creature faces a lone menace attacker
// has no legal block and no longer owes a decision: the window can be
// auto-passed because there is nothing in it for them. This used to
// err deliberately toward stopping, because a per-pair check could
// not see a count; now it does not have to guess.
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
	// One legal block is all it takes, so the generator is asked to
	// stop at the first one it finds rather than build the whole
	// option set for an answer this throws away — it runs on every
	// snapshot, for every seat.
	//
	// S27's planeswalker and battle cases live inside the generator:
	// "attacking this seat" is the DEFENDING player of the attack, not
	// a bare id comparison, which is what kept a player under a full
	// planeswalker assault from being told they owed no block
	// decision.
	return len(g.blockOptionsLocked(seat, 1, 1)) > 0
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
	// There is no legality pass here any more (#750, ADR 0045
	// addendum Decision 13). The block COUNTS this used to close out
	// — menace's minimum, and the maxima #750's block rules add — are
	// refused by DeclareBlockers when the declaration is made, so
	// every block that reaches the lock-in is already legal and the
	// lock-in has nothing to undo. That ordering is the fix: reverting
	// here meant the defender was told a lone block on a menace
	// attacker was good and only found out otherwise when the arrow
	// vanished.
	//
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
	// Pass 2: the attackers that are now BLOCKED (CR 509.1h) — which
	// is also, exactly, the attackers that "become blocked" and get
	// an EventBecomesBlocked (CR 506.4). One record for both: an
	// attacker becomes blocked once, and stays blocked for the rest
	// of the combat however few creatures are still blocking it. An
	// attacker already in the set does not become blocked a second
	// time when another creature is added to its block.
	if g.blockedAttackers == nil {
		g.blockedAttackers = map[uuid.UUID]bool{}
	}
	var blocked []pairing
	for _, p := range fresh {
		if g.blockedAttackers[p.attacker] {
			continue
		}
		g.blockedAttackers[p.attacker] = true
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

// clearBlockStateLocked forgets this combat's block declaration:
// what it announced, and which attackers it left blocked. Called from
// clearCombatLocked, alongside the BlockingTarget wipe they describe
// — next combat's identical pairing is a new declaration, announces
// again, and blocks again. Caller must hold g.mu.
func (g *Game) clearBlockStateLocked() {
	g.announcedBlocks = nil
	g.blockedAttackers = nil
}

// attackerBlockedLocked reports whether the attacker is blocked
// (CR 509.1h). The one reader of the blocked state outside the
// lock-in that writes it: the combat damage steps ask this, never the
// live blocker count, so a blocked attacker whose blockers have all
// died, been bounced or been removed from combat is still blocked and
// still assigns no damage to the player (CR 510.1c, #715).
//
// Caller must hold g.mu (read or write).
func (g *Game) attackerBlockedLocked(attacker uuid.UUID) bool {
	return g.blockedAttackers[attacker]
}
