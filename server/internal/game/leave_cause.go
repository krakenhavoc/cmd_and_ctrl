package game

import "github.com/google/uuid"

// leave_cause.go — WHY a permanent left the battlefield, on the
// zone-change event (#4395, ADR 0209 §3).
//
// Before this the event said where a permanent went and nothing about
// the route that took it there. A destruction, a lethal-damage death, a
// sacrifice, an exile and a bounce of one creature all emitted the same
// EventZoneMove, and the one engine fact that told them apart (the
// route template: destroyRoute, sbaDestroyRoute, battlefieldExitRoute,
// sacrificeRoute, exileRoute, bounceRoute) was gone by the time the
// event was written. The client needs the distinction to pick a look
// and a sound for the exit, so the route template now names it, once,
// and the two functions that perform every battlefield exit stamp it.
//
// # Where it is set
//
// Not per card. There are exactly two movers off the battlefield and
// each reads the cause off the route it was handed:
//
//   - executeBattlefieldLeaveLocked, the destroy / sacrifice / SBA exit.
//     routeBattlefieldExitInBatchLocked reads the template (leaveCauseOfExit)
//     and carries it on ReplacementEvent.LeaveCause, so it survives a
//     CR 614 prompt with everything else the move was asked for.
//   - executeZoneRouteLocked, the general effect route (exile, bounce,
//     tuck). leaveCauseOfRoute reads the route's destination.
//
// The sandbox's raw move (moveCardByRefLocked, to the stack) and the
// leaving-player cleanup stamp LeaveCauseOther by hand.
//
// # What the values mean
//
// The cause is the INSTRUCTION that moved the permanent, not where it
// landed. A destruction that a replacement (Rest in Peace) turns into an
// exile is still LeaveCauseDestroy, with NewZone exile; a client that
// wants the exile look reads NewZone first and the cause second. That
// keeps one rule for the engine ("name the verb") and leaves the
// destination to the field that already carries it.
//
//   - destroy: an effect that uses the word "destroy" (CR 701.8b).
//   - sba: a state-based action (CR 704.5). That is lethal damage and
//     deathtouch (704.5g, 704.5h) which CR 701.8b also calls a
//     destruction, and zero toughness (704.5f), zero loyalty (704.5i),
//     the legend rule (704.5j), an illegally attached Aura (704.5m) and
//     a battle at zero defense (704.5v, 704.5w). The client's "dies"
//     cue is this one; "destroy" is the cue for an effect.
//   - sacrifice: CR 701.21a, by effect or as a cost.
//   - exile_effect: a spell, ability or cost exiled it (CR 701.13).
//   - bounce: a spell, ability or cost returned it to a hand.
//   - other: everything else, a tuck into a library, a "put into a
//     graveyard" that is not a destruction, the sandbox, a player
//     leaving the game.

// LeaveCause is why a permanent left the battlefield. The empty value
// means "not a battlefield exit" (or nothing recorded).
type LeaveCause string

const (
	LeaveCauseDestroy   LeaveCause = "destroy"
	LeaveCauseSBA       LeaveCause = "sba"
	LeaveCauseSacrifice LeaveCause = "sacrifice"
	LeaveCauseExile     LeaveCause = "exile_effect"
	LeaveCauseBounce    LeaveCause = "bounce"
	LeaveCauseOther     LeaveCause = "other"
)

// leaveMark is what a battlefield exit stamps on its EventZoneMove: why,
// and the card behind it (Event.LeaveCause and Event.LeaveSource).
type leaveMark struct {
	cause  LeaveCause
	source uuid.UUID
}

// stamp writes the mark onto an event. A zero mark leaves it as it was.
func (m leaveMark) stamp(ev *Event) {
	ev.LeaveCause = m.cause
	ev.LeaveSource = m.source
}

// leaveCauseOfExit is the cause of a route that takes the destroy /
// sacrifice / SBA exit (ViaBattlefieldLeave), read off the template the
// verb chose. A destruction by an effect is a destroy; a sacrifice is a
// sacrifice; every other route on this exit is a state-based action,
// lethal-damage destruction (sbaDestroyRoute) included.
func leaveCauseOfExit(r zoneRoute) LeaveCause {
	switch {
	case r.Sacrifice:
		return LeaveCauseSacrifice
	case r.Destruction && !r.StateBased:
		return LeaveCauseDestroy
	}
	return LeaveCauseSBA
}

// leaveCauseOfRoute is the cause of a permanent leaving by the general
// effect route: what the route was asked to do with it. A manual
// sandbox move is nobody's instruction, so it is "other".
func leaveCauseOfRoute(r zoneRoute) LeaveCause {
	if r.Cause.Kind == MoveCauseManual {
		return LeaveCauseOther
	}
	switch r.Dst {
	case ZoneExile:
		return LeaveCauseExile
	case ZoneHand:
		return LeaveCauseBounce
	}
	return LeaveCauseOther
}

// leaveMoveCause is the cause a route's leave source is judged by: a
// route that settles now is a cost paid for a spell or ability
// (CR 602.2b, MustSettleNow), anything else is read from the resolving
// slot.
func (r zoneRoute) leaveMoveCause() MoveCause {
	if r.MustSettleNow {
		return MoveCause{Kind: MoveCauseCost}
	}
	return r.Cause
}

// leaveSourceLocked is the card behind a battlefield exit: the card
// whose spell or ability the event batch is the resolution of. It
// reads the resolving slot, which holds the item from the moment it
// begins to resolve until the next event batch opens (beginEventBatchLocked),
// so a state-based action that runs right after the resolution (a
// Blasphemous Act's lethal damage, a Toxic Deluge's -X/-X) is
// attributed to the spell that caused it, the same as an effect exit
// made while the item resolves. Every exit of one Damnation therefore
// shares one Batch and one LeaveSource.
//
// A cost, a special action and a sandbox move are nobody's resolution,
// and a state-based action with no resolution in the batch (combat
// damage) has no single source: all of them return uuid.Nil.
//
// Known looseness: a state-based action caused by something that is
// not a resolution, in the same batch as one that is (a planeswalker's
// loyalty cost taken after a spell resolved, with no step change
// between), reads the earlier item. The field groups; it does not
// prove.
//
// Caller must hold g.mu.
func (g *Game) leaveSourceLocked(c MoveCause) uuid.UUID {
	switch c.Kind {
	case MoveCauseCost, MoveCauseSpecialAction, MoveCauseManual:
		return uuid.Nil
	}
	if g.carriedCause != nil && g.carriedCause.Source != uuid.Nil {
		return g.carriedCause.Source
	}
	if r := g.resolving; r != nil && r.item != nil {
		return r.item.SourceCardID
	}
	return uuid.Nil
}
