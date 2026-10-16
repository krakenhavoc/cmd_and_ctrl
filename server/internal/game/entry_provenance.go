package game

import "github.com/google/uuid"

// entry_provenance.go — which ability put a permanent onto the
// battlefield (#3933).
//
// Kodama of the East Tree: "Whenever another permanent you control
// enters, if it wasn't put onto the battlefield with this ability, you
// may put a permanent card with equal or lesser mana value from your
// hand onto the battlefield." Without the intervening "if" (CR 603.4)
// one Kodama chains off its own puts until the hand is empty; with it,
// the entry its own ability made does not trigger that ability again.
// So an entry event has to name the ability that made it.
//
// Every effect entry runs while the item that made it is resolving, and
// Game.resolving names that item through every paused continuation of
// its resolution (resolving_item.go) — Kodama's put waits on a prompt.
// The same inference move_cause.go makes for a zone change's cause, so
// the ~40 entry callers need not thread the item through: the EventETB
// is stamped here from the resolving item, and only while its
// resolution is open (resolutionOpen), never in the window after it
// finished where Game.resolving is stale and a land may be played.
//
// What is stamped is the item's identity as an ABILITY: its source
// object (StackItem.SourceObject, CR 400.7 — a Kodama that left and came
// back is a new object with a new ability) and its catalog row
// (Params.Ability). A spell resolving into a permanent is not stamped
// (it is not "put onto the battlefield" by anything), and a spell's
// effect carries no catalog row, so neither can match an ability.

// stampEntryProvenanceLocked fills an entry's EventETB with the source
// object and catalog row of the ability that made it. It leaves #1320's
// Cause fields alone: an ETB has always carried them empty, and frozen
// snapshot fixtures hold those zero values.
// A no-op when nothing is resolving. Caller must hold g.mu.
func (g *Game) stampEntryProvenanceLocked(ev *Event) {
	if !g.resolutionOpen || g.resolving == nil || g.resolving.item == nil {
		return
	}
	it := g.resolving.item
	if it.SourceObject.ID != uuid.Nil {
		src := it.SourceObject
		ev.CauseSource = &src
	}
	if it.Params.Ability != nil {
		ref := *it.Params.Ability
		ev.CauseAbility = &ref
	}
}

// enteredWithThisAbility reports whether an EventETB names this
// trigger's own ability as the one that put the permanent onto the
// battlefield: the same source object and the same catalog row. A row
// that cannot be named never matches, and neither does an entry that
// names no ability.
func enteredWithThisAbility(ev Event, source *Card, t TriggeredAbility) bool {
	if ev.Kind != EventETB || ev.CauseSource == nil || ev.CauseAbility == nil || source == nil {
		return false
	}
	if *ev.CauseSource != (ObjectRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}) {
		return false
	}
	ref := TriggeredAbilityRef(t)
	return ref != nil && *ref == *ev.CauseAbility
}
