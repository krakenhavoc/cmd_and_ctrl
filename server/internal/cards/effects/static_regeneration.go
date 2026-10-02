package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// static_regeneration.go — CR 701.19b, "If this creature would be
// destroyed, regenerate it" (Clergy and Knight of the Holy Nimbus): a
// regeneration that is a static ability rather than a shield, so it
// replaces EVERY destruction of the permanent, with nothing spent.
//
// Every catalog static regeneration is built here, because each one must
// ask the regeneration gate (game.RegenerationAllowedForEffect, ADR 0108
// §2): "can't be regenerated" stops a static regeneration as surely as
// it stops a shield. static_regeneration_census_test.go fails a
// replacement anywhere in the catalog that regenerates without asking it.

// RegenerateIfThisWouldBeDestroyed is "If this creature would be
// destroyed, regenerate it." as a Spec replacement: instead of being
// destroyed, its damage is removed, its controller taps it and it is
// removed from combat (CR 701.19b). A sacrifice, the legend rule or zero
// toughness is not a destruction and is not replaced. Controlled by the
// permanent's controller, who orders it against any other replacement
// (CR 616.1).
func RegenerateIfThisWouldBeDestroyed() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return src != nil && g.RegenerationAllowedForEffect(ev, src.InstanceID)
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			g.RegenerateInsteadForEffect(ev)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			if src == nil {
				return uuid.Nil
			}
			return src.Controller
		},
		Label: "If this creature would be destroyed, regenerate it",
	}
}
