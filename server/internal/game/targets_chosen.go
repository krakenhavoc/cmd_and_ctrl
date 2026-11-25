package game

import "github.com/google/uuid"

// targets_chosen.go — #6228: EventTargetsChosen, "whenever a player
// chooses one or more targets" (Psychic Battle).
//
// Targets are chosen in two places: as a spell or ability is announced
// (CR 601.2c for a spell, CR 602.2b for an activated ability, CR 603.3d
// for a triggered ability as it is put on the stack), which is
// emitAnnouncedTargetsLocked, and when an effect changes them
// (CR 115.7), which is RetargetStackItemForEffect. Both already fan out
// EventBecomesTarget per target; this is the one event per choosing
// beside it. A copy's targets are copied, not chosen (CR 707.10): a
// copy whose controller "may choose new targets" chooses them through
// the change-of-targets path, which emits this.

// emitTargetsChosenLocked emits one EventTargetsChosen when `targets`
// names at least one object or player, and only while a permanent on
// the battlefield watches the kind, so a table without one keeps the
// event log it had before the kind existed. `causeName` is the name of
// the object whose effect changed the targets, empty for an
// announcement.
//
// Caller must hold g.mu.
func (g *Game) emitTargetsChosenLocked(chooser, source, itemID uuid.UUID, targets []TargetRef, causeName string) {
	if countRealTargets(targets) == 0 || !g.battlefieldWatchesLocked(EventTargetsChosen) {
		return
	}
	g.EmitEvent(Event{
		Kind:        EventTargetsChosen,
		Actor:       chooser,
		Source:      source,
		StackItemID: itemID,
		SourceName:  causeName,
	})
}

// objectNameLocked is the layered name of the object `id` names, in
// whatever zone holds it, or "" when it names nothing. Caller must hold
// g.mu.
func (g *Game) objectNameLocked(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	z, pos := g.locateCardLocked(id)
	if z == nil || pos < 0 || pos >= len(z.Cards) {
		return ""
	}
	return EffectiveOf(&z.Cards[pos]).Name
}
