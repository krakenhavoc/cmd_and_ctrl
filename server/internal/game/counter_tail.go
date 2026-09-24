package game

import (
	"errors"

	"github.com/google/uuid"
)

// counter_tail.go is life_tail.go's sibling for counters: the rest of
// an effect that is sequenced BEHIND a counter placement, and the
// guarantee that it runs when the counters actually land.
//
// Issue #1282. A counter placement runs the CR 614 window
// (RepEventCounter), and a window holding two different applicable
// replacements — a Doubling Season beside a Hardened Scales — queues
// the CR 616 ordering prompt and returns with NOTHING placed; the
// placement is owed to the resume, an action later. Until this file
// the placement had no continuation form, so a caller that did more
// work on the next line did it while the prompt was open:
//
//   - Amass (CR 701.47c): "then the Army you amassed deals damage equal
//     to its power" (Widespread Brutality) dealt the Army's PRE-amass
//     power.
//   - A Saga's entry lore counter (CR 714.3a): the chapter check read
//     the lore count before the counter was there, found nothing to
//     fire, and the resume placed the counter without ever asking
//     again — chapter I was simply lost.
//   - Earthbend (ADR 0081): the "then" of the sentence ran before the
//     land had its counters.
//
// The fix is the shape every other pausable primitive already has
// (CreateTokensThenForEffect, ChangePlayerLifeThenForEffect,
// MillToZoneThenForEffect, SacrificeAllThenForEffect): the entry point
// takes `then`, the event carries it across the pause, and the ONE
// settled exit — applyResolvedCounterThenLocked — runs it, from both
// the inline path and the CR 616 / "may" replacement resume.

// counterTail is what a counter placement owes its CALLER once the
// CR 614 window settles: the rest of the effect.
//
// Unexported engine plumbing. The catalog never builds one; it hands a
// `then` to AddCounterThenForEffect and the entry point fills this in
// before the pipeline runs, because a CR 616 prompt returns without
// placing anything and the resume has nothing else to go on.
type counterTail struct {
	// then runs once the placement has settled, with the delta the
	// window settled on — the doubled or incremented count for a
	// placement, negative for a removal — and zero for a placement
	// that was replaced away (CR 614.10) or whose target has left the
	// game.
	//
	// It takes the live *Game rather than capturing one, on the same
	// undo-safety contract every other continuation frame follows. It
	// runs with g.mu held and after the placement's own
	// EventCounterPlaced, so it may read the counters back or start the
	// next placement itself.
	then func(g *Game, placed int) error
}

// runCounterTailLocked runs a settled counter event's continuation
// exactly once. The tail is cleared ON THE EVENT, not through the
// pointer — lifeTail's idiom — so an undo snapshot's copy of the
// paused event (cloneReplacementResume copies the event by value)
// still carries it, and an undone-then-redone answer runs the rest of
// the effect again rather than skipping it.
//
// Every TERMINAL outcome goes through here: landed, replaced away,
// target gone. A pause is not terminal; the resume reaches this later
// through applyResolvedCounterThenLocked.
//
// Caller must hold g.mu.
func (g *Game) runCounterTailLocked(ev *ReplacementEvent, placed int) error {
	if ev == nil || ev.counterTail == nil || ev.counterTail.then == nil {
		return nil
	}
	then := ev.counterTail.then
	ev.counterTail = nil
	return then(g, placed)
}

// applyResolvedCounterThenLocked lands a settled RepEventCounter and
// then runs its continuation. The one settled exit that carries a
// tail: the inline path in AddCounterByThenForEffect and the CR 616 /
// "may" replacement resume in applyResolvedReplacementEventLocked both come
// through here, so a paused placement and an unpaused one cannot
// disagree about when the rest of the effect runs.
//
// A placement whose target has gone (ErrCardNotFound, or the player
// half's ErrPlayerNotFound / ErrPlayerEliminated) is still a terminal
// outcome: the continuation is told zero before the error comes back,
// exactly as applyResolvedLifeChangeLocked does.
//
// Caller must hold g.mu.
func (g *Game) applyResolvedCounterThenLocked(ev *ReplacementEvent) error {
	if err := g.applyResolvedCounterLocked(ev); err != nil {
		if tailErr := g.runCounterTailLocked(ev, 0); tailErr != nil {
			g.EmitEvent(Event{
				Kind:     EventEffectError,
				ErrorMsg: "counter continuation failed: " + tailErr.Error(),
			})
		}
		return err
	}
	return g.runCounterTailLocked(ev, ev.CounterDelta)
}

// AddCounterThenForEffect is AddCounterForEffect with a continuation:
// put `delta` `name` counters on `cardID` (a negative delta removes
// them), and once the placement has actually settled, run `then`.
//
// WHEN `then` RUNS. Immediately, before this returns, on the ordinary
// path. After the affected player answers the CR 616 ordering prompt
// when two different counter replacements apply to the placement (a
// Doubling Season and a Hardened Scales) — the case #1282 exists for,
// and the reason this is a continuation rather than the next line.
// Also immediately for a zero delta and for a placement replaced away
// entirely, with placed = 0: the sentence after "then" is not
// conditional on the counters having landed, and a caller waiting on
// it must be told either way.
//
// Anything the effect does AFTER the counters and that reads them —
// "then it deals damage equal to its power", "then check for lore
// chapters", "if it has N or more counters" — belongs in `then`. A
// caller with nothing after the placement passes nil, or keeps calling
// AddCounterForEffect, which is this with nil.
//
// Caller must hold g.mu.
func (g *Game) AddCounterThenForEffect(cardID uuid.UUID, name string, delta int, then func(g *Game, placed int) error) error {
	return g.AddCounterByThenForEffect(uuid.Nil, cardID, name, delta, then)
}

// AddCounterByThenForEffect is AddCounterThenForEffect with the
// CR 120.3d placer named (see AddCounterByForEffect). The one body
// every card-side counter placement goes through.
//
// Caller must hold g.mu.
func (g *Game) AddCounterByThenForEffect(placer, cardID uuid.UUID, name string, delta int, then func(g *Game, placed int) error) error {
	if delta == 0 {
		// Nothing to place and no event for a replacement to see — but
		// still a terminal outcome for a caller sequenced behind it.
		if then != nil {
			return then(g, 0)
		}
		return nil
	}
	if name == "" {
		return ErrInvalidParam
	}
	ev := &ReplacementEvent{
		Kind:          RepEventCounter,
		CounterTarget: cardID,
		CounterName:   name,
		CounterDelta:  delta,
		CounterPlacer: placer,
	}
	if then != nil {
		ev.counterTail = &counterTail{then: then}
	}
	out, err := g.applyReplacementsLocked(ev)
	if errors.Is(err, errReplacementPending) {
		// A prompt is queued and the resume owns the placement now.
		// NOTHING has been placed, and the continuation is still owed:
		// applyResolvedReplacementEventLocked runs it when the counters
		// land.
		return nil
	}
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		g.clearReplacementEventLocked(ev.ID)
		return err
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		// CR 614.10 with a null replacement: no counters. The rest of
		// the effect still runs, told zero.
		return g.runCounterTailLocked(ev, 0)
	}
	return g.applyResolvedCounterThenLocked(out)
}

// AddCounterMustSettleNowForEffect is AddCounterForEffect for a
// caller that CANNOT pause — a mana ability's PreRider (#1370) — and
// that needs the settled count back synchronously rather than through
// a continuation.
//
// CR 605.3b: activating a mana ability is one indivisible step with no
// priority window inside it, so a CR 616 ordering prompt (a Doubling
// Season beside a Hardened Scales both watching this placement) cannot
// be raised here any more than it can inside produce_mana.go's own
// window on the mana itself. The event is marked mustSettleNow, which
// forecloses errReplacementPending: the apply-loop applies the
// gathered order in place of asking (the same escape an eliminated
// chooser already takes, CR 616.1f), and an effect that would ask its
// own question is skipped un-applied — weaker than printed, never
// stronger, the posture every other mustSettleNow path takes.
//
// Returns the delta that actually landed — doubled, incremented, or
// zero for a placement replaced away entirely (CR 614.10) or whose
// target has left the game. A caller that computes its own output from
// the board (Empowered Autogenerator's ProducedFunc reading the
// counters back) wants the READ, not this return value, so most
// callers can ignore it; it exists for a caller that wants to know
// without a second lookup.
//
// Caller must hold g.mu.
func (g *Game) AddCounterMustSettleNowForEffect(cardID uuid.UUID, name string, delta int) (int, error) {
	return g.AddCounterByMustSettleNowForEffect(uuid.Nil, cardID, name, delta)
}

// AddCounterByMustSettleNowForEffect is
// AddCounterMustSettleNowForEffect with the CR 120.3d placer named —
// AddCounterByForEffect's mustSettleNow twin.
//
// Caller must hold g.mu.
func (g *Game) AddCounterByMustSettleNowForEffect(placer, cardID uuid.UUID, name string, delta int) (int, error) {
	if delta == 0 {
		return 0, nil
	}
	if name == "" {
		return 0, ErrInvalidParam
	}
	ev := &ReplacementEvent{
		Kind:          RepEventCounter,
		CounterTarget: cardID,
		CounterName:   name,
		CounterDelta:  delta,
		CounterPlacer: placer,
		mustSettleNow: true,
	}
	out, err := g.applyReplacementsLocked(ev)
	if err != nil && !errors.Is(err, ErrReplacementIterationExceeded) {
		// mustSettleNow forecloses errReplacementPending, so the only
		// way here is a genuinely broken pipeline. Land the counters
		// unreplaced rather than losing them — weaker than printed,
		// never stronger, matching replaceProducedManaLocked's own
		// fallback in produce_mana.go.
		g.clearReplacementEventLocked(ev.ID)
		if err := g.applyCounterByLocked(cardID, name, delta, placer, uuid.Nil); err != nil {
			return 0, err
		}
		return delta, nil
	}
	defer g.clearReplacementEventLocked(ev.ID)
	if out == nil || out.Canceled {
		return 0, nil
	}
	if err := g.applyResolvedCounterLocked(out); err != nil {
		return 0, err
	}
	return out.CounterDelta, nil
}
