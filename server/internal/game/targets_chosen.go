package game

import "github.com/google/uuid"

// targets_chosen.go — #6228: "Whenever a player chooses one or more
// targets" (Psychic Battle). EventTargetsChosen is emitted once per
// announcement, right after that announcement's EventBecomesTarget
// fan-out, from every place a player chooses targets:
//
//   - a cast, an activation and a triggered ability's placement
//     (emitAnnouncedTargetsLocked; CR 601.2c, 602.2b, 603.3d);
//   - a copy of a spell or ability, whose controller may choose new
//     targets (CR 707.10c);
//   - a change of targets (CR 115.7), stamped with the name of the card
//     whose effect changed them (Event.RetargetSourceName), so a card
//     that says its own change doesn't trigger it can tell.
//
// Gated on a battlefield watcher like EventTriggerPlaced, so a table
// without one keeps its event log.

// emitTargetsChosenLocked emits the EventTargetsChosen an announcement
// owes, when it named at least one object or player and a permanent on
// the battlefield watches the kind. `by` is the name of the card whose
// effect changed the targets, empty for a first announcement.
//
// Caller must hold g.mu.
func (g *Game) emitTargetsChosenLocked(actor, source, itemID uuid.UUID, targets []TargetRef, by string) {
	if countRealTargets(targets) == 0 || !g.battlefieldWatchesLocked(EventTargetsChosen) {
		return
	}
	g.EmitEvent(Event{
		Kind:               EventTargetsChosen,
		Actor:              actor,
		Source:             source,
		StackItemID:        itemID,
		RetargetSourceName: by,
	})
}
