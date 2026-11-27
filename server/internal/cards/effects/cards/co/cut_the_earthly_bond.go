package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cut the Earthly Bond — Instant — Arcane {U} (Saviors of Kamigawa, #5601):
//
//	"Return target enchanted permanent to its owner's hand."
//
// "Enchanted" is a permanent with an Aura attached to it (CR 303.4), judged as
// the spell is cast and again as it resolves (CR 608.2b), any controller's. The
// Aura need not be the one it is bouncing alongside.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "127ce189-2758-4789-9e02-4aa71d013431",
		Name:         "Cut the Earthly Bond",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetPermanent("target enchanted permanent", func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			return effects.IsEnchantedByAura(g, c.InstanceID)
		}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceChosenTarget(ctx.Game, item)
		},
	})
}
