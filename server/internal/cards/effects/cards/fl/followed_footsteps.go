package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Followed Footsteps — Enchantment — Aura {3}{U}{U} (Conspiracy: Take the
// Crown, S82 older sets batch 272, #5540):
//
//	"Enchant creature
//	 At the beginning of your upkeep, create a token that's a copy of
//	 enchanted creature."
//
// The creature is read as the trigger resolves: the token copies the
// copiable values of whatever the Aura enchants then (CR 707.2), under its
// controller's control, and nothing is made if the Aura is no longer attached
// to a creature. The copy may be of an opponent's creature; it is the Aura
// controller's token either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "732b1314-68bb-4d3d-9f3f-7e66e763caf5",
		Name:         "Followed Footsteps",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Followed Footsteps — create a token that's a copy of enchanted creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					host := g.AttachedHostOf(&src)
					if host == nil || !host.IsCreature() {
						return nil
					}
					return effects.CreateTokenCopy{Controller: ctx.Controller(), Copy: host.InstanceID, N: 1}.Apply(ctx)
				}),
		},
	})
}
