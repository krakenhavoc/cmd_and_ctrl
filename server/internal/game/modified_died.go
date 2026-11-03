package game

import "github.com/google/uuid"

// modified_died.go — "a modified creature died" (CR 700.9, #4834,
// Intermediate Chirography): whether a permanent that has just left the
// battlefield was modified as it last existed there (CR 603.10a). Read
// by the turn tally as a creature's death is counted
// (PlayerTurnTally.ModifiedCreaturesDied).

// departedWasModifiedLocked reports whether the card's most recent
// departure record says it was modified: it had a counter of any kind,
// an Equipment was attached to it, or an Aura controlled by its
// controller was. An attachment is read as the object the link named,
// live if it is still on the battlefield and from its own departure
// record if it left at the same time. A card with no departure record
// (a hand-built move in a test) was not modified.
//
// Caller must hold g.mu.
func (g *Game) departedWasModifiedLocked(cardID uuid.UUID) bool {
	info, ok := g.LastKnownPermanentForEffect(cardID)
	if !ok {
		return false
	}
	for _, n := range info.Counters {
		if n > 0 {
			return true
		}
	}
	for _, ref := range info.Attachments {
		var subtypes, types []string
		var controller uuid.UUID
		if a := findBattlefieldCard(g, ref.ID); a != nil && a.ObjectEpoch == ref.Epoch {
			e := a.Effective()
			subtypes, types, controller = e.Subtypes, e.Types, a.Controller
		} else if rec, ok := g.lastKnownPermanentLocked(ref); ok {
			subtypes, types, controller = rec.Characteristic.Subtypes, rec.Characteristic.Types, rec.Controller
		} else {
			continue
		}
		if hasTypeFold(subtypes, "Equipment") {
			return true
		}
		if hasTypeFold(types, "Enchantment") && hasTypeFold(subtypes, "Aura") && controller == info.Controller {
			return true
		}
	}
	return false
}
