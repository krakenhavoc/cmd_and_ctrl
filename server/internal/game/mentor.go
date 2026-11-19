package game

import "github.com/google/uuid"

// mentor.go — the event that says a creature mentored a creature (#5932).
// The keyword itself is built by the catalog (cards/effects Mentor); this
// file is the one thing the engine owns, so that another permanent can
// watch (Aegis of the Legion: "Whenever equipped creature mentors a
// creature, …").
//
//	702.134c An ability that triggers whenever a creature mentors another
//	creature triggers whenever a mentor ability whose source is the first
//	creature and whose target is the second creature resolves.
//
// So EventMentored is emitted once per mentor ability that resolves with
// its target still legal. A mentor ability whose target has become illegal
// does not resolve (CR 608.2b) and emits nothing. A +1/+1 counter that a
// replacement changed or that could not be placed does not matter: the
// rule asks only that the ability resolved.
//
// It is its own kind, as EventExploited is, rather than a predicate over
// EventCounterPlaced: that event does not say which ability placed the
// counter, and "mentors" is a fact about the mentor ability.

// EventMentored records that a creature mentored a creature (CR
// 702.134c). Source is the creature with mentor (the mentor ability's
// source), Actor the player who controlled the ability, and CardID the
// creature that was mentored (the ability's target), so the trigger's
// CR 603.10 object is that creature as it was mentored and a trigger that
// acts on "that creature" follows it only while it is the same object
// (CR 400.7).
const EventMentored EventKind = "mentored"

// MentoredForEffect records that `mentor`'s mentor ability, controlled by
// `controller`, resolved with `mentored` as its target (CR 702.134c). Call
// it from the mentor ability's resolution, once its target is known to be
// legal. Caller must hold g.mu in write mode.
func (g *Game) MentoredForEffect(mentor, controller, mentored uuid.UUID) {
	g.EmitEvent(Event{
		Kind:   EventMentored,
		Actor:  controller,
		CardID: mentored,
		Source: mentor,
	})
}
