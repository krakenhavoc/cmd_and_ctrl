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
// (Player.LosesAtNextSBA, or life at 0) that the next SBA pass reads,
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
// Idempotent: the cleanup hook runs it again after a discard pause
// drains, and a player leaving during that pause runs it once more.
//
// Must run BEFORE the cursor moves to the next turn: the turn-scoped
// statics and impulse grants compare their stamp against the current
// Turn.Number.
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
	// per-turn scope as DamageMarked, cleared at the same site.
	for i := range g.Battlefield.Cards {
		g.Battlefield.Cards[i].DamageMarked = 0
		g.Battlefield.Cards[i].MarkedLethalByDeathtouch = false
	}
	// S17 sub-PR 5: "until end of turn" replacement effects (Fog's
	// prevent-all-combat-damage, future prevention shields with a
	// per-turn duration) clear at cleanup so next turn starts with a
	// clean slate.
	g.ClearTurnScopedReplacementsLocked()
	// S32: "until end of turn" CONTINUOUS effects (Giant Growth's
	// +3/+3, Overrun's trample grant) expire here for the same reason
	// and by the same rule — CR 514.2 ends them during the cleanup
	// step, before the turn-based discard. This is also what makes a
	// grant created during the END step end this turn rather than
	// next: the sweep keys on the turn number stamped at registration,
	// not on "the next cleanup after the one I saw".
	g.ClearExpiredTurnScopedStaticsLocked()
	// S21 sub-PR 6: impulse-exile permissions ("you may play it this
	// turn") lapse here for the same reason — the turn they were
	// granted for is over. The exiled card stays exiled; it just stops
	// being playable.
	g.clearExpiredExilePlayLocked()
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
		// Wrap again from this seat's (never-taken) cleanup.
		skipped := next
		skipped.Step = StepCleanup
		next = skipped.advance(n)
	}
	g.Turn = next
	g.DiscardPending = nil
	g.onTurnBeganLocked()
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
	if g.LoyaltyActivatedThisTurn != nil {
		g.LoyaltyActivatedThisTurn = nil
	}
	if g.SpellsCastThisTurn != nil {
		g.SpellsCastThisTurn = nil
	}
	if g.LandsPlayedThisTurn != nil {
		g.LandsPlayedThisTurn = nil
	}
	if g.DrawnThisTurn != nil {
		g.DrawnThisTurn = nil
	}
	g.resetTurnTallyLocked()
}

// advancePastEliminatedLocked moves play on after a player has left
// the game. Called once per batch of departures, and only when the
// game goes on (settleDeparturesLocked).
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
// size is skipped too; the player it belonged to has left. The
// departed player's permanents stay on the battlefield (CR 800.4a is
// #769's).
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
