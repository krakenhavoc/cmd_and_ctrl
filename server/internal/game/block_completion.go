package game

import "github.com/google/uuid"

// block_completion.go gives the declare-blockers turn-based action a
// COMPLETION POINT per defending player (#1279, ADR 0045 Decision 38).
//
// The gap. CR 509.1 makes declaring blockers one turn-based action,
// taken as the step begins and before anyone receives priority. This
// engine keeps the sandbox's shape instead — entering the step hands
// the ACTIVE player priority, and a defender's block verbs are accepted
// while that window is open (blockers.go's header, #328) — so until
// this file nothing marked the declaration done. An empty blocked
// record meant "this defender chose not to block" and "this defender
// has not been asked yet" alike, and every reader that cared guessed:
// ninjutsu read an attacker as unblocked the instant the step began,
// the #328 auto-pass guard kept stopping a defender who had already
// declared, and the bot's block grace waited on "still has a legal
// block" rather than on the defender being done.
//
// The completion point. Each defending player's declaration is PENDING
// until one of four things completes it:
//
//  1. the step begins and the defender has no legal block at all —
//     "declared, none", with nothing to ask (autoCompleteBlockDeclarationsLocked);
//  2. the defender sends finish_blocks (FinishBlocks) — the explicit
//     "done blocking" / "no blocks" button, which needs no priority;
//  3. the defender PASSES PRIORITY in the step — how a table that
//     blocks by hand has always said "done", and still does;
//  4. the cursor leaves the step (the priority wrap, AdvanceStep) with
//     the defender still pending — completed as whatever is staged,
//     because the step cannot end on an unfinished turn-based action.
//
// Completing a declaration announces it: the defender's staged blocks
// are locked in (EventBlock, EventBecomesBlocked — #830's lock-in, now
// per defender) and then EventBlockersDeclared names the defender and
// how many creatures they blocked with. When the LAST pending defender
// completes through (2) or (3), the whole CR 509.1 action is done:
// state-based actions and the trigger drain run (CR 509.2) and the
// ACTIVE player receives priority (CR 117.3a), which is the window a
// ninjutsu player is owed and the sandbox used to skip.
//
// What stays permissive. A defender's block verbs are NOT refused after
// their declaration completes — a table that blocks by hand and passes
// a beat early can still put the block down, and #830's "a second
// blocker added after the lock-in announces its own block" still
// holds. What changes is that nothing OFFERS a block after completion:
// the option generator (blockOptionsLocked) answers empty for a
// declared defender, so the enumerator, the bot and the #328 signal all
// stop asking. The ADR records this as a sandbox allowance, not a rule.

// BlockDeclarationStatus is where one defending player's CR 509.1
// block declaration stands (#1279).
type BlockDeclarationStatus string

const (
	// BlockDeclarationNone — the seat is not a defending player this
	// combat, or the cursor is not at the declare-blockers step or a
	// later combat step, so there is no declaration to ask about.
	BlockDeclarationNone BlockDeclarationStatus = ""
	// BlockDeclarationPending — a defending player who has not
	// completed their declaration yet. Attackers pointed at them are
	// neither blocked nor unblocked.
	BlockDeclarationPending BlockDeclarationStatus = "pending"
	// BlockDeclarationDeclared — the declaration is complete, with or
	// without blocks. "Declared, none" is this status with no blocked
	// attacker, and it is the case the engine could not name before.
	BlockDeclarationDeclared BlockDeclarationStatus = "declared"
)

// BlockDeclarationStatusOf reports where `seat`'s block declaration
// stands. Takes the read lock; callers must not hold g.mu.
func (g *Game) BlockDeclarationStatusOf(seat uuid.UUID) BlockDeclarationStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.BlockDeclarationStatusLocked(seat)
}

// BlockDeclarationStatusLocked is BlockDeclarationStatusOf under a lock
// the caller already holds (read or write).
//
// In the combat steps AFTER declare_blockers every defending player
// reads Declared: the cursor cannot leave the step with a declaration
// still pending (completion point 4), so a pending bit found there is
// a board the sandbox built by hand, and "the declaration happened" is
// the only answer the damage steps can use.
func (g *Game) BlockDeclarationStatusLocked(seat uuid.UUID) BlockDeclarationStatus {
	if !g.blockersDeclaredStepLocked() || !g.isDefendingPlayerLocked(seat) {
		return BlockDeclarationNone
	}
	if g.Turn.Step != StepDeclareBlockers || g.blocksDeclared[seat] {
		return BlockDeclarationDeclared
	}
	return BlockDeclarationPending
}

// BlockDeclarationSeatsLocked returns the seat INDICES of the defending
// players whose declaration is pending, and of those whose declaration
// is complete, in seat order — the wire projection's
// block_pending_seats / blocks_declared_seats. Both are nil outside the
// declare-blockers step. A seat in neither list is not defending.
//
// Caller must hold g.mu (read or write).
func (g *Game) BlockDeclarationSeatsLocked() (pending, declared []int) {
	if g.Turn.Step != StepDeclareBlockers {
		return nil, nil
	}
	for i, s := range g.Seats {
		if s == nil {
			continue
		}
		switch g.BlockDeclarationStatusLocked(s.ID) {
		case BlockDeclarationPending:
			pending = append(pending, i)
		case BlockDeclarationDeclared:
			declared = append(declared, i)
		}
	}
	return pending, declared
}

// FinishBlocks completes `seat`'s block declaration (#1279): the
// "done blocking" / "no blocks" verb, finish_blocks on the wire.
// Whatever the seat has staged is its declaration; a seat that staged
// nothing has declared no blocks.
//
// Needs no priority — the declaration is a turn-based action, not
// something a player does with priority, and a defender must be able
// to finish while the active player still holds it (the same reason
// declare_blocker is not priority-gated, ADR 0033 §2).
//
// Errors: ErrGameNotActive, ErrWrongStep outside declare_blockers, a
// *ChoicePendingError while a blocking prompt is open (#730, the gate a
// pass obeys — completing may queue triggers and move priority),
// ErrPlayerNotFound for an unknown seat, and ErrNotDefending for a seat
// nothing is attacking. Idempotent: a seat whose declaration is already
// complete gets nil and nothing happens.
func (g *Game) FinishBlocks(seat uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.Turn.Step != StepDeclareBlockers {
		return ErrWrongStep
	}
	if c := g.blockingChoiceLocked(); c != nil {
		return choicePendingErrorLocked(c)
	}
	if p := g.playerByIDLocked(seat); p == nil || p.Eliminated {
		return ErrPlayerNotFound
	}
	g.RecomputeLayersIfStaleLocked()
	if !g.isDefendingPlayerLocked(seat) {
		return ErrNotDefending
	}
	if g.completeBlockDeclarationLocked(seat) {
		g.closeBlockDeclarationIfCompleteLocked()
	}
	return nil
}

// isDefendingPlayerLocked reports whether `seat` is a live player that
// some attacking creature's attack is defended by — the player
// attacked, the controller of the planeswalker attacked or the
// protector of the battle attacked, including one whose planeswalker or
// battle has since left (CR 506.4c, #1364). The same
// defendingPlayerForAttackerLocked read the block verb and the option
// generator use, so "who declares blockers" has one answer.
//
// Caller must hold g.mu (read or write).
func (g *Game) isDefendingPlayerLocked(seat uuid.UUID) bool {
	if seat == uuid.Nil || g.Battlefield == nil {
		return false
	}
	if p := g.playerByIDLocked(seat); p == nil || p.Eliminated {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget == uuid.Nil {
			continue
		}
		if g.defendingPlayerForAttackerLocked(c) == seat {
			return true
		}
	}
	return false
}

// defendingSeatsAPNAPLocked returns the defending players in APNAP
// order — starting after the active seat and walking the table — so
// the declarations a single boundary completes announce in the order
// CR 101.4 puts their triggers.
//
// Caller must hold g.mu (read or write).
func (g *Game) defendingSeatsAPNAPLocked() []uuid.UUID {
	n := len(g.Seats)
	if n == 0 {
		return nil
	}
	start := g.Turn.ActiveSeat
	if start < 0 || start >= n {
		start = 0
	}
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		s := g.Seats[(start+i)%n]
		if s == nil {
			continue
		}
		if g.isDefendingPlayerLocked(s.ID) {
			out = append(out, s.ID)
		}
	}
	return out
}

// allBlockDeclarationsCompleteLocked reports whether every defending
// player's declaration is complete — the CR 509.1 turn-based action as
// a whole. True with no defending players at all.
//
// Caller must hold g.mu (read or write).
func (g *Game) allBlockDeclarationsCompleteLocked() bool {
	for _, seat := range g.defendingSeatsAPNAPLocked() {
		if !g.blocksDeclared[seat] {
			return false
		}
	}
	return true
}

// blockDeclarationCompleteForAttackerLocked reports whether the
// declaration that decides `attacker`'s blocked-or-not has happened:
// its defending player's, or — for an attacker no live player defends
// — every defender's. Outside the declare-blockers step it answers the
// step question alone (blockersDeclaredStepLocked).
//
// Caller must hold g.mu (read or write).
func (g *Game) blockDeclarationCompleteForAttackerLocked(attacker *Card) bool {
	if !g.blockersDeclaredStepLocked() {
		return false
	}
	if g.Turn.Step != StepDeclareBlockers {
		return true
	}
	if d := g.defendingPlayerForAttackerLocked(attacker); d != uuid.Nil {
		return g.blocksDeclared[d]
	}
	return g.allBlockDeclarationsCompleteLocked()
}

// blockAnnounceableLocked reports whether a block staged by `blocker`
// may be announced by the lock-in: its controller's declaration is
// complete. Outside the declare-blockers step nothing gates it (the
// verb cannot stage there, and a block found there was built by hand).
//
// Caller must hold g.mu (read or write).
func (g *Game) blockAnnounceableLocked(blocker *Card) bool {
	if g.Turn.Step != StepDeclareBlockers {
		return true
	}
	return g.blocksDeclared[blocker.Controller]
}

// completeBlockDeclarationLocked completes `seat`'s block declaration:
// marks it, locks in its staged blocks, and emits EventBlockersDeclared
// with the number of creatures it blocked with. Reports whether it
// completed anything — false outside the declare-blockers step, for a
// seat that is not defending, and for one already complete, so every
// completion point can call it unconditionally.
//
// Triggers the announcement produces are queued, not drained; the
// caller decides when a player next receives priority.
//
// Caller must hold g.mu in write mode.
func (g *Game) completeBlockDeclarationLocked(seat uuid.UUID) bool {
	if g.State != StateActive || g.Turn.Step != StepDeclareBlockers {
		return false
	}
	if g.blocksDeclared[seat] || !g.isDefendingPlayerLocked(seat) {
		return false
	}
	if g.blocksDeclared == nil {
		g.blocksDeclared = map[uuid.UUID]bool{}
	}
	g.blocksDeclared[seat] = true
	// The defender's staged blocks announce first, so the blocked
	// record is written before anything reading EventBlockersDeclared
	// asks it.
	g.commitBlockDeclarationLocked()
	blockers := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.BlockingTarget != uuid.Nil && c.Controller == seat {
			blockers++
		}
	}
	g.EmitEvent(Event{
		Kind:   EventBlockersDeclared,
		Actor:  seat,
		Amount: blockers,
	})
	return true
}

// completeAllBlockDeclarationsLocked completes every pending defender's
// declaration, in APNAP order — completion point 4, run as the cursor
// leaves the step. Reports whether it completed any.
//
// Caller must hold g.mu in write mode.
func (g *Game) completeAllBlockDeclarationsLocked() bool {
	if g.Turn.Step != StepDeclareBlockers {
		return false
	}
	completed := false
	for _, seat := range g.defendingSeatsAPNAPLocked() {
		if g.completeBlockDeclarationLocked(seat) {
			completed = true
		}
	}
	return completed
}

// autoCompleteBlockDeclarationsLocked is completion point 1, run as the
// declare-blockers step begins: every defending player with no legal
// block has declared none. There is nothing to ask them, so there is no
// reason for ninjutsu, an "attacks and isn't blocked" trigger or the
// bot's grace to wait on them.
//
// Layers must be fresh (the option generator reads evasion off the
// effective characteristics); finishStepEntryLocked guarantees it.
//
// Caller must hold g.mu in write mode.
func (g *Game) autoCompleteBlockDeclarationsLocked() {
	for _, seat := range g.defendingSeatsAPNAPLocked() {
		if len(g.blockOptionsLocked(seat, 1, 1)) == 0 {
			g.completeBlockDeclarationLocked(seat)
		}
	}
}

// closeBlockDeclarationIfCompleteLocked ends the CR 509.1 turn-based
// action once every defender has declared: CR 509.2's triggers go on
// the stack at the priority boundary (runStateChecksLocked), and the
// ACTIVE player receives priority (CR 117.3a). Reports whether it did.
//
// Called after a completion by pass or by finish_blocks. Not after
// completion point 4 — that one is the step ending, and the callers
// there run their own boundary.
//
// Caller must hold g.mu in write mode.
func (g *Game) closeBlockDeclarationIfCompleteLocked() bool {
	if !g.allBlockDeclarationsCompleteLocked() {
		return false
	}
	g.runStateChecksLocked()
	if g.State == StateActive {
		g.Turn.PriorityHolder = g.Turn.ActiveSeat
	}
	return true
}
