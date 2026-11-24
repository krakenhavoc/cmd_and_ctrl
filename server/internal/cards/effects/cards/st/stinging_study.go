package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stinging Study — Instant {4}{B} (Secret Lair Drop):
//
//	"You draw X cards and you lose X life, where X is the mana value of a
//	 commander you own on the battlefield or in the command zone."
//
// X is read as the spell resolves over the commanders the caster owns on the
// battlefield or in the command zone (Majestic Genesis's read); with two
// commanders it is the greater, since nothing is gained by the smaller. With
// none, X is zero.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "115bc79e-a109-42d7-ac58-2ee8dc0a959f",
		Name:         "Stinging Study",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			x := effects.GreatestOwnedCommanderManaValue(ctx.Game, item.Controller)
			if x <= 0 {
				return nil
			}
			if err := (effects.DrawCards{N: x}).Apply(ctx); err != nil {
				return err
			}
			return effects.GainLife{Amount: -x}.Apply(ctx)
		},
	})
}
