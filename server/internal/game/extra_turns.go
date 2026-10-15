package game

import "github.com/google/uuid"

// extra_turns.go — CR 500.7 extra turns (ADR 0059 Decision 5, #753).
//
// "Take an extra turn after this one" puts a turn directly after the
// current one. Several of them are taken one at a time, and the most
// recently created is taken first (CR 500.7). So the queue is a STACK:
// TakeExtraTurnsForEffect pushes onto the end of Game.ExtraTurns and
// the rotation seam (beginNextTurnLocked, rotation.go) pops from the
// end. A Time Warp cast during the first of Time Stretch's two turns
// is therefore taken before the second, with no rule of its own.
//
// An extra turn is a turn. It gets its own Turn.Seq, the seat's
// TurnsBegun goes up, and every per-turn reset runs as it begins
// (onTurnBeganLocked), so a planeswalker can be activated again, the
// land drop comes back and "until your next turn" ends. Turn.Round
// does not move: "Turn N" on the board is the round (ADR 0059 owner
// decision 1), and Turn.Extra marks the turn instead.
//
// When the queue is empty, normal rotation resumes from Turn.OrderSeat
// — the seat whose NORMAL turn this is, or last was — so an extra turn
// seat 0 gives seat 2 in a four-player game is followed by seat 1's
// normal turn, not seat 3's.
//
// A queued turn of a player who has left the game doesn't begin
// (CR 800.4k). It is dropped as it would begin, and it still counts
// toward that seat's TurnsBegun (CR 800.4m), exactly as a departed
// seat's passed-over normal turn does.

// ExtraTurn is one queued CR 500.7 extra turn.
type ExtraTurn struct {
	// Ref is the turn's identity, minted from Game.NextExtraRef. It
	// is stable for the life of the queue entry and becomes Turn.ExtraRef
	// when the turn begins, so a delayed trigger can be bound to "that
	// turn" (DelayedTrigger.OnExtraTurn: Final Fortune's end-step loss).
	Ref int `json:"ref"`
	// Seat is the seat index that takes the turn.
	Seat int `json:"seat"`
	// Source is the card whose effect created the turn, for the log.
	Source uuid.UUID `json:"source,omitempty"`
	// Effects are rules effects (no affected objects) created by the
	// effect that queued this turn, built and validated when the card
	// resolved (DuringExtraTurnRuleEffectForEffect) and registered as the
	// turn begins with UntilEndOfTurnDuration() and their original
	// Timestamp (ADR 0188 Decision 3b): Alchemist's Gambit's "During that
	// turn, damage can't be prevented". They live on the queue entry, so
	// a turn that is skipped or dropped takes them with it (CR 614.10a)
	// and an undo that restores the queue restores them.
	Effects []ScopedEffect `json:"effects,omitempty"`
}

// TakeExtraTurnsForEffect gives `player` n extra turns directly after
// the current turn (CR 500.7). It returns their refs in the order they
// will be taken — the first one taken first — and emits one
// EventExtraTurnAdded per turn.
//
// Returns nil, and queues nothing, for n < 1, for a player who is not
// seated and for one who has left the game: an effect that tells a
// departed player to take a turn does nothing (CR 800.4a).
//
// Safe to call from inside a resolution: it only queues. The turn
// begins at the rotation seam once this one ends, never here.
//
// Caller must hold g.mu.
func (g *Game) TakeExtraTurnsForEffect(player, source uuid.UUID, n int) []int {
	if n < 1 {
		return nil
	}
	p := g.playerByIDLocked(player)
	if p == nil || p.Eliminated {
		return nil
	}
	refs := make([]int, n)
	for i := range refs {
		g.NextExtraRef++
		refs[i] = g.NextExtraRef
	}
	// The stack is popped from the end, so the turn taken first is
	// pushed LAST: all n of this effect's turns go ahead of every turn
	// queued earlier (CR 500.7, "most recently created first"), and
	// among themselves they come out in ref order.
	//
	// A fresh slice every time, never an in-place append: RestoreFrom
	// hands the live game the backing array an undo snapshot holds,
	// and writing into its spare capacity would rewrite that snapshot
	// (the trap sweepScopedStaticsLocked documents).
	queue := make([]ExtraTurn, len(g.ExtraTurns), len(g.ExtraTurns)+n)
	copy(queue, g.ExtraTurns)
	for i := n - 1; i >= 0; i-- {
		queue = append(queue, ExtraTurn{Ref: refs[i], Seat: p.Seat, Source: source})
	}
	g.ExtraTurns = queue
	for _, ref := range refs {
		g.EmitEvent(Event{
			Kind:   EventExtraTurnAdded,
			Actor:  player,
			Source: source,
			Amount: ref,
		})
	}
	return append([]int(nil), refs...)
}

// queueAnchoredExtraTurnLocked queues an extra turn anchored to a turn
// that is not necessarily the current one when the effect was created:
// Emrakul, the Promised End's "After that turn, that player takes an
// extra turn" (ADR 0167 §3). ADR 0059 Decision 5 left this anchor for
// the first card that named another turn.
//
// The caller (endTurnControlsLocked) runs while "that turn" is still
// g.Turn, so the turn goes directly after it (CR 500.7). The extra
// turns created during that turn are "after this one" too and were
// created later, so they are taken first: they are the entries whose
// Ref is above floor (NextExtraRef as that turn began), at the top of
// the stack. The new entry goes directly below them, and every other
// entry stays below it.
//
// Like TakeExtraTurnsForEffect it builds a fresh slice and never writes
// into the old one, which an undo snapshot may share. It emits
// EventExtraTurnAdded, so the extra_turn log line and the view's
// extra_turns need nothing new. A seat out of range or out of the game
// gets nothing (CR 800.4k).
//
// Caller must hold g.mu.
func (g *Game) queueAnchoredExtraTurnLocked(seat int, source uuid.UUID, ref, floor int) {
	if seat < 0 || seat >= len(g.Seats) || g.Seats[seat] == nil || g.Seats[seat].Eliminated {
		return
	}
	at := len(g.ExtraTurns)
	for at > 0 && g.ExtraTurns[at-1].Ref > floor {
		at--
	}
	queue := make([]ExtraTurn, 0, len(g.ExtraTurns)+1)
	queue = append(queue, g.ExtraTurns[:at]...)
	queue = append(queue, ExtraTurn{Ref: ref, Seat: seat, Source: source})
	queue = append(queue, g.ExtraTurns[at:]...)
	g.ExtraTurns = queue
	g.EmitEvent(Event{
		Kind:   EventExtraTurnAdded,
		Actor:  g.Seats[seat].ID,
		Source: source,
		Amount: ref,
	})
}

// ExtraTurnsQueuedForEffect is a copy of the extra-turn queue in the
// order the turns will be taken, next first. Caller must hold g.mu.
func (g *Game) ExtraTurnsQueuedForEffect() []ExtraTurn {
	return g.extraTurnsInOrderLocked()
}

// ExtraTurnsQueued is the locking form of ExtraTurnsQueuedForEffect.
func (g *Game) ExtraTurnsQueued() []ExtraTurn {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.extraTurnsInOrderLocked()
}

// extraTurnsInOrderLocked reverses the stack into take order.
func (g *Game) extraTurnsInOrderLocked() []ExtraTurn {
	if len(g.ExtraTurns) == 0 {
		return nil
	}
	out := make([]ExtraTurn, 0, len(g.ExtraTurns))
	for i := len(g.ExtraTurns) - 1; i >= 0; i-- {
		out = append(out, g.ExtraTurns[i])
	}
	return out
}

// popExtraTurnLocked takes the next queued extra turn of a seat still
// in the game. Entries of seats that have left are dropped on the way
// (CR 800.4k) and counted toward their TurnsBegun (CR 800.4m). ok is
// false when the queue held no playable turn.
//
// paused is true when the candidate turn's player owes a skip AND an
// extra-turn skip applies to it too (ADR 0188 Decision 12, case 3): the
// entry is back on top of the queue, Game.TurnChangePending is set, and
// a turn_skip_order prompt asks the player which skip takes it.
//
// Caller must hold g.mu.
func (g *Game) popExtraTurnLocked() (et ExtraTurn, ok, paused bool) {
	for len(g.ExtraTurns) > 0 {
		last := len(g.ExtraTurns) - 1
		et := g.ExtraTurns[last]
		// Re-sliced to a length the caller can only shrink or copy
		// from: TakeExtraTurnsForEffect never appends in place, so the
		// shared backing array is never written through this slice.
		g.ExtraTurns = g.ExtraTurns[:last:last]
		if len(g.ExtraTurns) == 0 {
			g.ExtraTurns = nil
		}
		if et.Seat < 0 || et.Seat >= len(g.Seats) || g.Seats[et.Seat] == nil {
			continue
		}
		if g.Seats[et.Seat].Eliminated {
			g.noteTurnBegunLocked(et.Seat)
			// CR 800.4k: a departed player's skip debt can never be spent.
			g.dropTurnSkipsOfSeatLocked(et.Seat)
			continue
		}
		// ADR 0188 Decision 1: a player who owes a skip skips the next
		// turn they would begin, and this is that turn (the Magosi
		// ruling). Decision 12's three cases:
		//
		//  1. no skip owed: the RepEventExtraTurn window runs, below;
		//  2. a skip owed and no extra-turn skip applies: the charge is
		//     spent, with no question;
		//  3. a skip owed and an extra-turn skip (Trouble in Pairs,
		//     Ugin's Nexus) applies too: both are replacements on one
		//     event and CR 616.1 lets the player choose which takes the
		//     turn. The turn change pauses to ask.
		if _, _, owes := g.oldestTurnSkipLocked(et.Seat); owes {
			if others := g.extraTurnSkipAppliesLocked(et); len(others) > 0 && g.pauseForTurnSkipOrderLocked(et, others) {
				return ExtraTurn{}, false, true
			}
			source, _ := g.consumeTurnSkipLocked(et.Seat)
			g.emitTurnSkippedLocked(et.Seat, source, et.Ref)
			continue
		}
		if g.extraTurnSkippedLocked(et) {
			continue
		}
		return et, true, false
	}
	return ExtraTurn{}, false, false
}

// extraTurnSkippedLocked opens the CR 614 window over a queued extra
// turn that is about to begin and reports whether it settled on a skip
// (CR 614.10): "if an opponent would begin an extra turn, that player
// skips that turn instead" (Trouble in Pairs, #2529).
//
// A skipped turn never begins. It is already off the queue, so nothing
// bound to it can fire — sweepUnreachableBoundTriggersLocked drops a
// delayed trigger tied to its Ref as the next turn starts (CR 614.10a:
// "anything scheduled for a skipped turn won't happen", which is why
// Final Fortune's lose-the-game never comes) — and it does not count
// toward the seat's TurnsBegun, the same call ADR 0059 Decision 1 makes
// for every turn skipped by an effect: that count is the turns that
// BEGAN. A seat that has left the game is dropped before this runs
// (CR 800.4k), and the window never sees it.
//
// The window cannot pause (mustSettleNow): the rotation seam's one
// resume point is ADR 0188 Decision 12's which-skip question
// (turn_skip_order.go), asked before this window opens, and a CR 616
// ordering inside the window has no resume. See RepEventExtraTurn.
//
// Caller must hold g.mu.
func (g *Game) extraTurnSkippedLocked(et ExtraTurn) bool {
	ev := &ReplacementEvent{
		Kind:          RepEventExtraTurn,
		Actor:         g.Seats[et.Seat].ID,
		Source:        et.Source,
		ExtraTurnSeat: et.Seat,
		ExtraTurnRef:  et.Ref,
		mustSettleNow: true,
	}
	out, err := g.applyReplacementsLocked(ev)
	defer g.clearReplacementEventLocked(ev.ID)
	// A cancelled event comes back as (nil, nil) — the signal
	// runStepEntryHooksLocked reads the same way. An error is a broken
	// pipeline (mustSettleNow forecloses a pause), and the turn is then
	// taken: weaker than printed for the skipper, never a turn that
	// silently vanishes.
	if err != nil || (out != nil && !out.Canceled) {
		return false
	}
	g.EmitEvent(Event{
		Kind:   EventExtraTurnSkipped,
		Actor:  ev.Actor,
		Source: et.Source,
		Amount: et.Ref,
	})
	return true
}

// extraTurnPendingLocked reports whether `ref` names the turn now in
// progress or one still queued. A delayed trigger bound to any other
// ref can never fire and is swept (sweepUnreachableBoundTriggersLocked).
func (g *Game) extraTurnPendingLocked(ref int) bool {
	if g.Turn.ExtraRef == ref {
		return true
	}
	for _, et := range g.ExtraTurns {
		if et.Ref == ref {
			return true
		}
	}
	return false
}

// extraTurnQueuedLocked reports whether `ref` names an extra turn still
// on the queue (not the one in progress).
func (g *Game) extraTurnQueuedLocked(ref int) bool {
	for _, et := range g.ExtraTurns {
		if et.Ref == ref {
			return true
		}
	}
	return false
}

// DuringExtraTurnRuleEffectForEffect attaches a rules effect to queued
// extra turn `ref`: "During that turn, damage can't be prevented"
// (Alchemist's Gambit; ADR 0188 Decision 3b, CR 611.2a). The effect is
// created now but starts with the turn, because no Duration begins in a
// future turn and starting it now would cover the rest of this turn too.
//
// scope must not be ScopeNone: an effect that modifies objects fixes its
// affected set when it begins (CR 611.2c), which a deferred record cannot
// honour, so the API takes a scope and never affected objects. The record
// is validated now by the checks every registration makes
// (buildScopedEffectLocked) and stamped with the creation timestamp
// (CR 613.7a). `controller` is the effect's "you", as for
// RegisterScopedRuleEffectForEffect.
//
// Returns false, writing nothing, when ref is not queued, the scope is
// unusable or there are no mods. Copy on write: the queue is rebuilt, so
// an undo snapshot that shares the old backing array is not touched.
//
// Caller must hold g.mu (write).
func (g *Game) DuringExtraTurnRuleEffectForEffect(ref int, sourceID uuid.UUID, scope AffectedScope, controller uuid.UUID, mods []Mod, label string) bool {
	if scope == ScopeNone || !KnownAffectedScope(scope) || len(mods) == 0 {
		return false
	}
	at := -1
	for i, et := range g.ExtraTurns {
		if et.Ref == ref {
			at = i
			break
		}
	}
	if at < 0 {
		return false
	}
	// The duration is a placeholder: startTurnLocked stamps
	// UntilEndOfTurnDuration() as the turn begins.
	e := g.buildScopedEffectLocked(sourceID, nil, scope, controller, mods, IndefiniteDuration(), label, timeNowUnixNano())
	queue := cloneExtraTurns(g.ExtraTurns)
	queue[at].Effects = append(queue[at].Effects, e)
	g.ExtraTurns = queue
	return true
}

// registerExtraTurnEffectsLocked puts the deferred effects of the extra
// turn that has just begun on the registry, each lasting until the end of
// this turn (CR 514.2) and keeping the timestamp it was created with
// (CR 613.7a). Called from startTurnLocked after onTurnBeganLocked, so
// the turn-began sweeps cannot see them.
//
// Caller must hold g.mu (write).
func (g *Game) registerExtraTurnEffectsLocked(effects []ScopedEffect) {
	for _, e := range effects {
		e.Duration = g.UntilEndOfTurnDuration()
		e.Mods = cloneMods(e.Mods)
		g.appendBuiltScopedEffectLocked(e)
	}
}

// cloneExtraTurns deep-copies the queue: the deferred Effects carry
// slices of their own (Mods, Affected), and an undo snapshot must not
// share them with the live game.
func cloneExtraTurns(in []ExtraTurn) []ExtraTurn {
	if len(in) == 0 {
		return nil
	}
	out := make([]ExtraTurn, len(in))
	for i, et := range in {
		if len(et.Effects) > 0 {
			effects := make([]ScopedEffect, len(et.Effects))
			for j, e := range et.Effects {
				e.Affected = append([]AffectedObject(nil), e.Affected...)
				e.Mods = cloneMods(e.Mods)
				effects[j] = e
			}
			et.Effects = effects
		}
		out[i] = et
	}
	return out
}
