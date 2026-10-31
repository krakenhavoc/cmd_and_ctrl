package game

import "github.com/google/uuid"

// connive.go — the CR 701.50 keyword action "<permanent> connives"
// (#4641). The draw, the discard and the +1/+1 counter are built by the
// catalog (cards/effects/connive.go); this file is the one thing the
// engine owns: the event that says the permanent has connived.
//
//	701.50a … that permanent's controller draws a card, then discards a
//	card. If a nonland card is discarded this way, that player puts a
//	+1/+1 counter on the conniving permanent.
//	701.50f A permanent "connives" after the process … is complete, even
//	if some or all of those actions were impossible.
//
// So EventConnived is emitted once per conniving permanent, after its
// discard and its counter, on every branch: an empty hand (nothing to
// discard) and a permanent that has left the battlefield (CR 701.50b, the
// draw and discard still happen, only the counter has nowhere to go)
// included. Several permanents conniving together (CR 701.50c) connive one
// after another, so each is its own event and "whenever a creature you
// control connives" triggers once per creature. CR 701.50e: connive 0 is
// no connive and emits nothing; the callers return before the draw.
//
// It is its own kind, as EventExplored is, so a connive payoff fires on
// nothing else (a plain draw and discard is not a connive).

// EventConnived records that a permanent connived (CR 701.50f). CardID is
// the permanent that connived and Actor the player who controlled it as it
// connived (its last controller, when it has left the battlefield; CR
// 701.50b), so "a creature you control connives" is Actor == the ability's
// controller and holds for a permanent that is no longer there. Source is
// the permanent itself.
const EventConnived EventKind = "connived"

// ConnivedForEffect records that `creature` has finished conniving under
// `controller` (CR 701.50f). Caller must hold g.mu.
func (g *Game) ConnivedForEffect(creature, controller uuid.UUID) {
	g.EmitEvent(Event{
		Kind:   EventConnived,
		Actor:  controller,
		CardID: creature,
		Source: creature,
	})
}
