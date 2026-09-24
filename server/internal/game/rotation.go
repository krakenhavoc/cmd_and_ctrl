package game

// rotation.go is the single turn-rotation seam (ADR 0059 Decision 6,
// #766). Every way a turn can end goes through it:
//
//   - the cursor walking past cleanup (advanceCursorLocked),
//   - the active player leaving the game (advancePastEliminatedLocked,
//     reached from Concede and from the SBA loss pass),
//   - the sandbox pass_turn verb (PassTurn).
//
// Before this seam the last two wrote the next seat's Untap into
// g.Turn by hand. The departed player's turn never ran its cleanup
// sweep, the per-turn caches were not reset, and its attackers stayed
// in combat — so an attacker of a player who had conceded hit the next
// player again in that player's own combat damage step (#766).
//
// The constraint every caller honours: beginNextTurnLocked and the
// step entry hooks that follow it never run inside a resolving
// callback, a replacement's Apply, or a prompt continuation that is
// still part of a resolution. Beginning a turn untaps, resets the
// per-turn state, announces steps and queues upkeep triggers, and none
// of that may happen in the middle of a spell. The callers today are
// the cursor advance, an SBA pass and a player action (Concede,
// PassTurn), which are all action boundaries. Nothing in a resolution
// eliminates a player directly: a loss during a resolution sets a flag
// (Player.AttemptedEmptyDraw, or life at 0) that the next SBA pass reads,
// and that pass is the resolution bookend's or a later prompt
// answer's. ADR 0057's effect losses keep the same shape
// (Game.ActiveSeatLeftPending, consumed by the SBA loss pass); a new
// caller that can be reached mid-resolution must defer the same way.

// sweepTurnEndLocked is the non-interactive part of the cleanup step
// (CR 514.2): marked damage and deathtouch marks are removed, and
// "until end of turn" and "this turn" effects end. It was the body of
// the StepCleanup entry hook, moved out unchanged so a turn that ends
// early runs exactly the same sweep.
//
// Every "this turn" registry sweeps here. When ADR 0057's
// TurnScopedGameEndGates or ADR 0045's TurnScopedBlockRules land,
// their sweeps join this function rather than the cleanup hook.
//
// Idempotent: CR 514.3a's second cleanup step runs it again (#661),
// and a player leaving during a discard pause runs it once more.
//
// Not safe in the middle of a state-based action pass: the lethal
// damage and deathtouch SBAs (CR 704.5g/h) read the marks this clears.
// A player who loses in an SBA pass leaves at once, but the turn does
// not end until repeated SBA passes have settled. A lord dying in one
// pass can make another creature lethally damaged on the next pass;
// runStateChecksLocked waits for that before rotating.
//
// Must run BEFORE the cursor moves to the next turn: the impulse
// grants compare their stamp against the current Turn.Number, and an
// "until end of turn" continuous effect ends at the cleanup step of
// the turn it was made in, not at the start of the next one.
//
// Caller must hold g.mu.
func (g *Game) sweepTurnEndLocked() {
	// CR 514.2: damage marked on permanents is removed. The
	// lethal-damage SBA from S13.1 reads DamageMarked, so clearing it
	// here means the per-turn damage doesn't carry over into the next
	// turn.
	//
	// S18 sub-PR 3: MarkedLethalByDeathtouch is the companion flag
	// (CR 702.2c) set by combat damage from deathtouch sources. Same
	// per-turn scope as DamageMarked, cleared at the same site — and,
	// since #816, by the same helper the battlefield exit uses, so the
	// two places that clear marked damage cannot disagree about what
	// clearing means.
	for i := range g.Battlefield.Cards {
		clearBattlefieldDamage(&g.Battlefield.Cards[i])
	}
	// #667 / CR 701.19a: a regeneration shield lasts until it is used
	// or until the turn ends, and this is the second of those. Swept
	// beside the marked damage because they are the same kind of
	// thing — per-turn state on one permanent — and because a shield
	// that outlived its turn would save a creature next turn from a
	// destruction nobody paid for.
	g.clearRegenerationShieldsLocked()
	// S17 sub-PR 5: "until end of turn" replacement effects (Fog's
	// prevent-all-combat-damage, future prevention shields with a
	// per-turn duration) clear at cleanup so next turn starts with a
	// clean slate.
	g.ClearTurnScopedReplacementsLocked()
	// #750, ADR 0045 addendum Decision 11: until-end-of-turn BLOCK
	// rules ("this creature can't block this turn") end here for the
	// same reason and by the same rule. The registry is emptied
	// wholesale — it holds nothing that outlasts a turn — and it sits
	// beside the replacement sweep rather than in the layer duration
	// sweep below because a block rule is not a continuous effect the
	// layers apply; it is a question asked when a block is declared.
	g.ClearTurnScopedBlockRulesLocked()
	// S32: "until end of turn" CONTINUOUS effects (Giant Growth's
	// +3/+3, Overrun's trample grant) expire here for the same reason
	// and by the same rule — CR 514.2 ends them during the cleanup
	// step, before the turn-based discard. This is also what makes a
	// grant created during the END step end this turn rather than
	// next: the end step is not the end of the turn, so the cleanup
	// sweep such a grant meets is this turn's own.
	//
	// S38 (ADR 0063): the same sweep now walks every CR 611.2
	// duration, and the argument is what tells `durationExpiredLocked`
	// that the current moment is a cleanup step. A "for as long as ~"
	// or "until your next turn" effect passing through here is not
	// ended by this turn being over — one function decides, and for
	// those two it says no.
	g.ClearEndOfTurnScopedStaticsLocked()
	// S21 sub-PR 6, ADR 0066: granted cast and play permissions ("you
	// may play it this turn") lapse here for the same reason — the
	// turn they were granted for is over. The exiled card stays
	// exiled; it just stops being playable. Hygiene rather than
	// correctness: an expired permission is already refused.
	//
	// #945: through the same ADR 0063 Duration and the same
	// durationExpiredLocked the statics above just ran, with the same
	// `true` — this moment is a cleanup step, and that is the only
	// thing the sweep has to tell it.
	g.sweepCastPermissionsLocked(true)
	// #663: an event-conditioned delayed trigger is "this turn" —
	// "when you NEXT cast an instant or sorcery spell THIS TURN" —
	// and CR 514.2 ends it here whether or not the cast it was
	// waiting for ever happened. Step-conditioned triggers carry no
	// duration and are untouched.
	g.clearExpiredDelayedTriggersLocked(true)
}

// beginNextTurnLocked picks the next turn and stamps the cursor on
// that seat's Untap, then runs the turn-began hook. It does not run
// the step entry hooks: callers do, as they did before the seam
// existed (the cursor advance's caller, advancePastEliminatedLocked,
// PassTurn).
//
// The next turn belongs to the next seat after the current active seat
// that is still in the game. A seat that has left is passed over (CR
// 800.4k: that player's turn doesn't begin), and Turn.Number still
// goes up when the rotation passes seat 0, so it keeps counting rounds
// whether or not seat 0 is still playing.
//
// A cleanup-discard pause cannot outlive its turn: the discard belongs
// to the turn that ended, so DiscardPending is dropped. In ordinary
// play it is already empty here (the cleanup hook advances only once
// it drains); the early ways out of a turn are what can leave it set.
//
// Caller must hold g.mu.
func (g *Game) beginNextTurnLocked() {
	n := len(g.Seats)
	if n == 0 {
		return
	}
	from := g.Turn
	from.Step = StepCleanup
	next := from.advance(n)
	for i := 0; i < n && next.ActiveSeat >= 0 && next.ActiveSeat < n && g.Seats[next.ActiveSeat].Eliminated; i++ {
		// CR 800.4m: an effect that lasts "until that player's next
		// turn" lasts until the turn that WOULD have begun. Counting
		// the skipped seat's never-taken turn here is what ends such
		// an effect at the right moment instead of leaving it live
		// for the rest of the game (ADR 0063 Decision 3).
		g.noteTurnBegunLocked(next.ActiveSeat)
		// Wrap again from this seat's (never-taken) cleanup.
		skipped := next
		skipped.Step = StepCleanup
		next = skipped.advance(n)
	}
	g.Turn = next
	g.DiscardPending = nil
	g.noteTurnBegunLocked(next.ActiveSeat)
	g.onTurnBeganLocked()
}

// noteTurnBegunLocked bumps `Player.TurnsBegun` for the seat whose
// turn is beginning (ADR 0059 Decision 1, implemented here for
// ADR 0063 / #755).
//
// It is the counter "until your next turn" ends on, and the reason it
// exists rather than arithmetic on `Turn.Number` is that Turn.Number
// counts ROUNDS: all four seats in a Commander game share one number,
// so "your next turn" cannot be expressed with it.
//
// Caller must hold g.mu.
func (g *Game) noteTurnBegunLocked(seat int) {
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil {
		return
	}
	g.Seats[seat].TurnsBegun++
}

// onTurnBeganLocked clears the per-turn caches as a turn begins. It
// runs from beginNextTurnLocked only, so it runs exactly once per turn
// and needs no "did the seat change?" test: before the rotation seam
// it compared seats, which skipped the resets whenever a seat's turn
// followed a turn of the same seat.
//
// Flushes `LoyaltyActivatedThisTurn` (CR 606.3 — once per turn per
// planeswalker), the spell, land and draw tallies, and the TurnTally.
// It is the home for any other "reset on new turn" cache the engine
// grows. Caller must hold g.mu.
func (g *Game) onTurnBeganLocked() {
	// S25 (#77): invalidate the layer cache. A continuous effect
	// whose AppliesTo reads the TURN rather than the battlefield —
	// Zurgo Helmsmasher's "during your turn, ~ has indestructible" is
	// the first in the catalog — changes its answer here and nowhere
	// else, so nothing would otherwise mark the cached
	// characteristics stale and the keyword would stick around on the
	// wrong player's turn.
	//
	// This is the bump layer_listener.go's header predicted and
	// deliberately deferred ("Step / phase advance … when they
	// arrive in a later sprint, advance the version inside the
	// step-advance helper directly — no event for it today"). It
	// lands here rather than in the listener for the reason that note
	// gives: there is no event for a turn change to listen to.
	//
	// Cost is one recompute per turn, against a cache that is already
	// invalidated by every zone move and every counter placed.
	g.layerVersion.Add(1)
	// S38 (ADR 0063 Decision 2): "until your next turn" ends as that
	// player's turn BEGINS. This hook runs from beginNextTurnLocked
	// after the cursor is stamped on the new seat's untap step and
	// before the step entry hooks untap anything, which is exactly
	// CR 500.1 + CR 502.1: the turn has begun, and the untapping
	// happens inside it. Putting the sweep in the untap turn-based
	// action instead would order it against "doesn't untap" effects
	// for no reason (ADR 0058).
	g.ClearExpiredScopedStaticsLocked()
	// #945: and the granted cast permissions, for the same reason and
	// at the same boundary. "Until the end of your next turn" is the
	// clause that needs it — Reckless Impulse's grant is stamped
	// against a turn that has not happened yet, so the cleanup sweep
	// of every turn before it leaves the grant alone and this is the
	// backstop for a turn that ended without one (ADR 0059
	// Decision 6).
	g.sweepCastPermissionsLocked(false)
	if g.LoyaltyActivatedThisTurn != nil {
		g.LoyaltyActivatedThisTurn = nil
	}
	if g.SpellsCastThisTurn != nil {
		g.SpellsCastThisTurn = nil
	}
	if g.ForetoldThisTurn != nil {
		g.ForetoldThisTurn = nil
	}
	if g.LandsPlayedThisTurn != nil {
		g.LandsPlayedThisTurn = nil
	}
	// #500: one-turn "play an additional land this turn" grants
	// expire here with the tally they were raising the ceiling over.
	// A permanent's standing grant is derived from the battlefield
	// every time it's asked, so it has nothing to reset.
	if g.ExtraLandDropsThisTurn != nil {
		g.ExtraLandDropsThisTurn = nil
	}
	if g.DrawnThisTurn != nil {
		g.DrawnThisTurn = nil
	}
	g.resetTurnTallyLocked()
	// #1255: last-known information for spells that left the stack
	// this turn. Nothing can still name one — the stack is empty at a
	// turn boundary, and every copy effect that names a spell was on it.
	g.clearLastKnownStackLocked()
	// #1379: the same argument for permanents that left the
	// battlefield — whatever referred to one was on the stack.
	g.clearLastKnownPermanentsLocked()
	// #1181: the per-turn half of the activation record dies with the
	// turn it counted. The game-lifetime half does not — "activate
	// each exhaust ability only once" is a claim about the whole game,
	// which is the reason the two scopes are separate maps.
	g.resetActivationTurnTallyLocked()
}

// advancePastEliminatedLocked moves play on after a player has left
// the game. Called once per batch of departures, and only when the
// game goes on: from settleDeparturesLocked (Concede), or from
// runStateChecksLocked after repeated state-based action passes settle.
//
// If the active seat is still in the game, the only work is priority:
// a holder who has left hands it back to the active seat (priority
// always defaults to active at step boundaries). NoPriority (Untap,
// Cleanup) is left alone.
//
// If the active seat has left, the rest of its turn ends at once: the
// attackers and blockers leave combat, the cleanup sweep runs, and the
// next turn begins through the rotation seam, then its step entry
// hooks run (the auto-untap walks the cursor on to Upkeep).
//
// That is a declared simplification of CR 800.4j, which says the turn
// "continues to its completion without an active player" (ADR 0059
// Decision 6, owner decision 2026-09-17). The steps that turn had left
// do not happen, so the other players' "at the beginning of each end
// step" triggers skip that one turn, and delayed triggers scheduled
// for its end step wait for the next player's. The discard to hand
// size is skipped too; the player it belonged to has left. Their
// objects are already gone by the time this runs — CR 800.4a is
// performed inside leaveGameLocked, not here (#769, leave_game.go).
//
// When the departure comes from an SBA pass, all repeated checks have
// destroyed what they had to before this runs. Those deaths are counted
// in the ended turn's TurnTally, which the new turn then resets, and
// their dies triggers wait on PendingTriggers and go on the stack in
// the next player's upkeep. That is part of the same simplification:
// under CR 800.4j they would go on the stack in the turn that is still
// running.
//
// See the file comment for why this never runs inside a resolution.
// Caller must hold g.mu.
func (g *Game) advancePastEliminatedLocked() {
	numSeats := len(g.Seats)
	if numSeats == 0 {
		return
	}
	active := g.Turn.ActiveSeat
	if active < 0 || active >= numSeats || !g.Seats[active].Eliminated {
		ph := g.Turn.PriorityHolder
		if ph >= 0 && ph < numSeats && g.Seats[ph].Eliminated {
			g.Turn.PriorityHolder = active
		}
		return
	}
	g.clearCombatLocked()
	g.sweepTurnEndLocked()
	g.beginNextTurnLocked()
	g.runStepEntryHooksLocked()
}
