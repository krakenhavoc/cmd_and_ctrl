package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// UmbraArmor is CR 702.89a — "If enchanted permanent would be
// destroyed, instead remove all damage marked on it and destroy this
// Aura" — as a Spec replacement on the Aura carrying it (printed "totem
// armor" on older cards, CR 702.89b).
//
// It replaces a destruction only: a destroy spell, a wipe, lethal or
// deathtouch damage. Sacrifice, exile, bounce and toughness 0 or less
// are not destructions and pass through. It is not a regeneration, so
// "can't be regenerated" does not stop it. The enchanted permanent's
// controller orders it against a regeneration shield or another umbra
// (CR 616.1); two copies are not interchangeable (each destroys its own
// Aura), so it declares ActsOnItsSource. The rule lives in
// game/umbra_armor.go.
func UmbraArmor() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return src != nil && g.UmbraArmorAppliesForEffect(ev, src.InstanceID)
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			if src == nil {
				return nil
			}
			return g.UmbraArmorReplaceForEffect(ev, src.InstanceID)
		},
		Controller: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) uuid.UUID {
			return g.AffectedPermanentControllerForEffect(ev)
		},
		ActsOnItsSource: true,
		Label:           "Umbra armor",
	}
}
