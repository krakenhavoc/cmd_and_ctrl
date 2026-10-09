package game

import "github.com/google/uuid"

// block_completion.go gives the declare-blockers turn-based action a
// COMPLETION POINT per defending player (#1279, ADR 0045 Decision 38).
//
// The gap. CR 509.1 makes declaring blockers one turn-based action,
// taken as the step begins and before anyone receives priority. This
// engine used to keep the sandbox's shape instead — entering the step
// handed the ACTIVE player priority, and a defender's block verbs were
// accepted while that window was open (blockers.go's header, #328) — so
// until this file nothing marked the declaration done. An empty blocked
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
//  3. the defender PASSES PRIORITY in the step — the manual table's
//     "done", for the one case a pending defender can still hold
//     priority (see below);
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
// Nobody has priority before the declaration (#1501). As the step
// begins, once (1) has run, priority is PARKED — Turn.PriorityHolder is
// NoPriority — while any defender is still declaring
// (beginBlockDeclarationLocked). Nobody can cast, activate or pass; a
// defender declares with the block verbs and finishes with
// finish_blocks, and the last one to finish hands the active player
// priority, as above. The cleanup step's discard parks priority the
// same way (cleanup.go). Everything that can leave the table parked
// with nobody left to wait for hands priority on: completion point 4
// (completeAllBlockDeclarationsLocked), a pending defender conceding
// (Concede), and — for every other path, a sandbox clear_combat or a
// mana ability that sacrificed the last attacker — the dispatcher's
// SettleBlockDeclaration after each action. Point (3) remains for a
// pending defender who holds priority anyway: a restore point written
// before #1501 mid-step.
//
// The declaration happens once (#2021). When every player defending at
// that moment has completed theirs, the CR 509.1 action as a whole is
// over and Game.blockDeclarationClosed records it. A player who becomes
// a defending player afterwards — an attack reselected onto them
// (CR 508.7a), a creature put onto the battlefield attacking them
// (CR 508.4) — was not defending when the turn-based action happened,
// so they never declare: they read as declared, are offered no block,
// are refused one (blocks_declared), and their pass is an ordinary
// pass. An attacker pointed at them stays as the declaration left it,
// blocked or unblocked (CR 509.1h).
//
// And a declaration that is complete is complete (#1501). The block
// verbs refuse a block from a defender whose declaration is done
// (blocks_declared, checkBlockRestrictionsLocked), so a ninja that
// entered attacking after the declaration cannot be blocked — an
// engine fact, where #1279 left it a table convention and recorded the
// late block as a sandbox allowance. The option generator
// (blockOptionsLocked) already answered empty for a declared defender;
// the verb now agrees with it.

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
	if g.Turn.Step != StepDeclareBlockers || g.blockDeclarationDoneLocked(seat) {
		return BlockDeclarationDeclared
	}
	return BlockDeclarationPending
}

// blockDeclarationDoneLocked reports whether `seat` can no longer
// declare blockers this combat: their own declaration is complete, or
// the declaration as a whole is over (#2021) — which answers for a
// player who became a defending player after it ended and so never
// made one. Every "has this defender declared" question in the engine
// asks this, never blocksDeclared directly.
//
// Caller must hold g.mu (read or write).
func (g *Game) blockDeclarationDoneLocked(seat uuid.UUID) bool {
	return g.blockDeclarationClosed || g.blocksDeclared[seat]
}

// noteBlockDeclarationClosedIfCompleteLocked records that the CR 509.1
// declaration as a whole is over once every player defending right now
// has completed theirs (#2021). Called wherever a declaration can be
// the last one — each completion, the step's entry with nobody to
// wait for, the step's exit, and the settle after an action — so the
// record never lags the moment the active player could receive
// priority. Never cleared within the step; clearBlockStateLocked
// clears it with the rest of combat.
//
// Caller must hold g.mu in write mode.
func (g *Game) noteBlockDeclarationClosedIfCompleteLocked() {
	if g.blockDeclarationClosed || g.State != StateActive || g.Turn.Step != StepDeclareBlockers {
		return
	}
	if g.allBlockDeclarationsCompleteLocked() {
		g.blockDeclarationClosed = true
	}
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
// something a player does with priority, and since #1501 nobody holds
// priority while it is being made (the same reason declare_blocker is
// not priority-gated, ADR 0033 §2). It is how a defender finishes: the
// last one to finish hands the active player priority.
//
// Errors: ErrGameNotActive, ErrWrongStep outside declare_blockers, a
// *ChoicePendingError while a blocking prompt is open (#730, the gate a
// pass obeys — completing may queue triggers and move priority),
// ErrPlayerNotFound for an unknown seat, ErrNotDefending for a seat
// nothing is attacking, and — #1597 — a *BlockRefusedError with reason
// block_requirement while the seat's staged declaration leaves a
// CR 509.1c requirement unobeyed that it could obey. Idempotent: a seat whose declaration is already
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
	// #1597 / CR 509.1c: finishing is the declaration's checkpoint,
	// exactly as the defender's pass is.
	if err := g.blockCheckpointLocked(seat); err != nil {
		return err
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

// IsDefendingPlayerForEffect reports whether `seat` is a defending
// player in the sense a spell or ability cast during combat means —
// Yare's and Blaze of Glory's "target creature defending player
// controls" (#1715). Commander plays with the attack-multiple-players
// option, under which "all the attacking player's opponents are
// defending players during the combat phase" from the moment it
// starts (CR 802.2), and a two-player game says the same of its one
// nonactive player (CR 506.2). So: any live player other than the
// active player, during the combat phase — which is what gives Blaze
// of Glory a target in the beginning of combat step, before anything
// attacks. Outside the combat phase there is no defending player.
//
// Not the block verb's question: WHO MAY BLOCK a given attacker is
// isDefendingPlayerLocked's, the player that attack is aimed at.
//
// Caller must hold g.mu (read or write).
func (g *Game) IsDefendingPlayerForEffect(seat uuid.UUID) bool {
	if PhaseOf(g.Turn.Step) != PhaseCombat || seat == uuid.Nil || seat == g.activeSeatIDLocked() {
		return false
	}
	p := g.playerByIDLocked(seat)
	return p != nil && !p.Eliminated
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
// a whole. True with no defending players at all, and true once the
// declaration has closed (#2021), whoever is defending since.
//
// Caller must hold g.mu (read or write).
func (g *Game) allBlockDeclarationsCompleteLocked() bool {
	if g.blockDeclarationClosed {
		return true
	}
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
		return g.blockDeclarationDoneLocked(d)
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
	return g.blockDeclarationDoneLocked(blocker.Controller)
}

// completeBlockDeclarationLocked completes `seat`'s block declaration:
// marks it, locks in its staged blocks, and emits EventBlockersDeclared
// with the number of creatures it blocked with. Reports whether it
// completed anything — false outside the declare-blockers step, for a
// seat that is not defending, for one already complete, and for one
// who became a defending player after the declaration closed (#2021):
// they never declared, so nothing is announced for them. Every
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
	if g.blockDeclarationDoneLocked(seat) || !g.isDefendingPlayerLocked(seat) {
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
	g.noteBlockDeclarationClosedIfCompleteLocked()
	return true
}

// completeAllBlockDeclarationsLocked completes every pending defender's
// declaration, in APNAP order — completion point 4, run as the cursor
// leaves the step. Reports whether it completed any.
//
// #1501: priority parked for the declaration comes back to the active
// player here, so AdvanceStep's CR 117.4 drive has a holder to pass
// from. The callers run their own boundary (state checks, the trigger
// drain); this only names who holds priority across it.
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
	g.noteBlockDeclarationClosedIfCompleteLocked()
	g.unparkBlockPriorityLocked()
	return completed
}

// beginBlockDeclarationLocked is the declare-blockers step's turn-based
// action as the step begins (CR 509.1), run by the step-entry hook:
// completion point 1 for every defender with nothing to decide, and
// then — #1501 — priority PARKED while anyone is still declaring.
// Nobody receives priority until the declaration is over; the last
// defender's finish_blocks hands it to the active player
// (closeBlockDeclarationIfCompleteLocked). With every defender already
// complete (none of them had a legal block) the active player keeps
// the priority the cursor gave them, exactly as before.
//
// Caller must hold g.mu in write mode, with fresh layers.
func (g *Game) beginBlockDeclarationLocked() {
	g.autoCompleteBlockDeclarationsLocked()
	// #2021: with nobody left declaring (or nobody defending at all)
	// the declaration is over as the step begins.
	g.noteBlockDeclarationClosedIfCompleteLocked()
	if !g.allBlockDeclarationsCompleteLocked() {
		g.grantPriorityLocked(NoPriority)
	}
}

// autoCompleteBlockDeclarationsLocked is completion point 1, run as the
// declare-blockers step begins: every defending player with no legal
// block has declared none. There is nothing to ask them, so there is no
// reason for ninjutsu, an "attacks and isn't blocked" trigger or the
// bot's grace to wait on them — and, since #1501, no reason for the
// table to wait on them before anyone receives priority.
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

// blockPriorityParkedLocked reports whether priority is parked for the
// block declaration (#1501): the game is running, the cursor is in
// declare_blockers, and nobody holds priority. The only way into that
// state is beginBlockDeclarationLocked.
//
// Caller must hold g.mu (read or write).
func (g *Game) blockPriorityParkedLocked() bool {
	return g.State == StateActive && g.Turn.Step == StepDeclareBlockers && g.Turn.PriorityHolder == NoPriority
}

// unparkBlockPriorityLocked hands parked priority to the active player
// once nobody is left declaring. It moves the holder and nothing else:
// the caller runs whatever boundary it owes.
//
// Caller must hold g.mu in write mode.
func (g *Game) unparkBlockPriorityLocked() {
	if g.blockPriorityParkedLocked() && g.allBlockDeclarationsCompleteLocked() {
		g.grantPriorityLocked(g.Turn.ActiveSeat)
	}
}

// SettleBlockDeclaration ends a block declaration that has nobody left
// to wait for while priority is still parked for it (#1501), with the
// same boundary the last finish_blocks runs: state-based actions, the
// trigger drain, and priority to the active player. A no-op in every
// other state, and cheap: the parked check is three field reads.
//
// The dispatcher calls it after every action (actions.Dispatch), beside
// SettleResolution. The engine's own paths out of a parked step settle
// themselves — FinishBlocks, Concede, AdvanceStep — so this is the
// belt for the rest: a sandbox clear_combat, a manual move of the last
// attacker, a mana ability whose sacrifice cost took it. Any of those
// can leave no defending player pending, and a parked step with nobody
// declaring would hold the table with no move for anyone.
func (g *Game) SettleBlockDeclaration() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.settleBlockDeclarationLocked()
}

// settleBlockDeclarationLocked is SettleBlockDeclaration under the
// caller's lock.
//
// Caller must hold g.mu in write mode.
func (g *Game) settleBlockDeclarationLocked() {
	if g.blockPriorityParkedLocked() {
		g.closeBlockDeclarationIfCompleteLocked()
	}
}

// closeBlockDeclarationIfCompleteLocked ends the CR 509.1 turn-based
// action once every defender has declared: CR 509.2's triggers go on
// the stack at the priority boundary (runStateChecksLocked), and the
// ACTIVE player receives priority (CR 117.3a). Reports whether it did.
//
// Called after a completion by pass or by finish_blocks, and by
// settleBlockDeclarationLocked when priority is parked with nobody
// left declaring (#1501). Not after completion point 4 — that one is
// the step ending, and the callers there run their own boundary.
//
// Caller must hold g.mu in write mode.
func (g *Game) closeBlockDeclarationIfCompleteLocked() bool {
	if !g.allBlockDeclarationsCompleteLocked() {
		return false
	}
	g.noteBlockDeclarationClosedIfCompleteLocked()
	g.runStateChecksLocked()
	if g.State == StateActive {
		g.grantPriorityLocked(g.Turn.ActiveSeat)
	}
	return true
}
