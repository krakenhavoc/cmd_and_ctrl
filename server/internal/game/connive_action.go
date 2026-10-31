package game

import "github.com/google/uuid"

// connive_action.go — connive as an event a replacement effect can apply to
// (CR 701.50a, CR 614.1a, #4589), and the first replacement of one: Leader,
// Super-Genius.
//
//	701.50a ... To do so, that permanent's controller draws a card, then
//	discards a card. If a nonland card is discarded this way, that player
//	puts a +1/+1 counter on the conniving permanent.
//	614.1a Effects that use the word "instead" are replacement effects.
//	614.5 A replacement effect doesn't invoke itself repeatedly; it gets
//	only one opportunity to affect an event or any modified events that
//	may replace that event.
//	614.6 If an event is replaced, it never happens. A modified event
//	occurs instead ...
//
// # The shape
//
// "Connive" is one instruction on one permanent, so it is one event: opened
// once at the one entry point (ConniveForEffect) as a RepEventKeywordAction
// with KeywordActionConnive, the way learn is, and performed only once the
// CR 614 window has settled. Leader says "If a creature you control would
// connive, instead you draw a card, then that creature connives." The
// modified event is a draw followed by a connive that this Leader has
// already had its one opportunity on (CR 614.5), so the replacement is a
// plain Replace on the event: it adds a card drawn first
// (ReplacementEvent.ConniveExtraDraws) and sets the count to 1, since "that
// creature connives" carries no "N". Two Leaders are each applied once to
// the modified event (CR 616.1f), so they draw two cards before the one
// connive.
//
// # The caller's tail
//
// The draw, the discard and the counter are the catalog's (the discard
// continuation differs per card: Cyclonus converts after the counter), so
// the caller hands in a KeyedTail, a registered body and its data, which
// the window runs once it has settled. It runs with Tail.N set to the
// connive's settled count, the extra draws as the outcome's Amount and the
// conniving permanent as its only ID. A connive that did not happen (a
// null replacement, CR 614.10, or a count of 0, CR 701.50e) runs the same
// tail with N left at 0: the rest of the sentence the connive was part of
// (the next creature of a chain, Cyclonus's conversion) still runs. The
// event carries data only, so a window that waits on a CR 616 ordering
// prompt is a restore point (replacement_resume_keyed.go).

// ConniveForEffect is "connive" for `creature` under `controller` (its
// controller as it connives, CR 701.50b), as `n` (1 for a plain connive).
// It opens the CR 614 window and runs `then` when it has settled; see the
// file comment for what `then` is told.
//
// Caller must hold g.mu (it is an effect-time helper).
func (g *Game) ConniveForEffect(controller, creature uuid.UUID, n int, then KeyedTail) error {
	if n <= 0 {
		// CR 701.50e: no connive event occurs.
		return g.runKeyedTailLocked(then, []uuid.UUID{creature}, 0)
	}
	then.N = 0
	_, err := g.runKeywordActionLocked(&ReplacementEvent{
		Kind:               RepEventKeywordAction,
		Actor:              controller,
		Source:             creature,
		KeywordAction:      KeywordActionConnive,
		KeywordActionCount: n,
		keywordAction:      &keywordActionTail{keyed: then},
	})
	return err
}

// applyConniveLocked runs the caller's tail for a connive the window left
// (possibly rewritten).
//
// Caller must hold g.mu.
func (g *Game) applyConniveLocked(ev *ReplacementEvent) error {
	tail := ev.keywordAction
	if tail == nil {
		return nil
	}
	// Taken as a value and cleared off the tail, as earthbend does: the
	// abandoned path also runs it, and only one of them may.
	keyed := tail.keyed
	tail.keyed = KeyedTail{}
	keyed.N = ev.KeywordActionCount
	return g.runKeyedTailLocked(keyed, []uuid.UUID{ev.Source}, ev.ConniveExtraDraws)
}
