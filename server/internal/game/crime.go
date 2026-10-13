package game

import "github.com/google/uuid"

// emitAnnouncedTargetsLocked is for a completed cast, activation, or
// triggered-ability placement (CR 700.13). Copies and retargeting call
// emitBecameTargetLocked directly: they choose targets without making a
// qualifying announcement. Caller holds g.mu.
func (g *Game) emitAnnouncedTargetsLocked(actor, source, itemID uuid.UUID, targets []TargetRef) {
	crimeSlot := -1
	for i, target := range targets {
		if g.crimeTargetLocked(actor, target) {
			crimeSlot = i
			break
		}
	}
	g.emitBecameTargetLocked(actor, source, itemID, targets, crimeSlot)
}

func (g *Game) crimeTargetLocked(actor uuid.UUID, target TargetRef) bool {
	if actor == uuid.Nil || target.ID == uuid.Nil {
		return false
	}
	switch target.Kind {
	case TargetPlayer:
		return target.ID != actor && g.playerByIDLocked(target.ID) != nil
	case TargetCard:
		// Abilities have stack metadata but no card in a zone. Spells
		// share their card ID, so this check covers both without guessing
		// the controller from the ability's source permanent.
		if item := g.StackMeta[target.ID]; item != nil {
			return item.Controller != uuid.Nil && item.Controller != actor
		}
		zone := g.findCardZoneLocked(target.ID)
		if zone == nil {
			return false
		}
		if zone.Kind == ZoneGraveyard {
			return zone.Owner != uuid.Nil && zone.Owner != actor
		}
		if zone.Kind == ZoneBattlefield || zone.Kind == ZoneStack {
			for i := range zone.Cards {
				if card := &zone.Cards[i]; card.InstanceID == target.ID {
					return card.Controller != uuid.Nil && card.Controller != actor
				}
			}
		}
	}
	return false
}
