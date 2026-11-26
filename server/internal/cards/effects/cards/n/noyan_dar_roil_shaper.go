package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Noyan Dar, Roil Shaper — Legendary Creature — Merfolk Ally {3}{W}{U}, 4/4
// (Battle for Zendikar):
//
//	"Whenever you cast an instant or sorcery spell, you may put three +1/+1
//	 counters on target land you control. If you do, that land becomes a 0/0
//	 Elemental creature with haste that's still a land."
//
// The counters and the animation are awaken's own (CR 702.113a,
// game.AwakenForEffect): the land is still a land, becomes a 0/0 Elemental
// creature with haste, and stays one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e4190320-5854-479e-be11-a5d5ef3db6a6",
		Name:         "Noyan Dar, Roil Shaper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Noyan Dar, Roil Shaper — you may put three +1/+1 counters on target land you control; it becomes a 0/0 Elemental creature with haste",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return g.AwakenForEffect(ctx.Controller(), ctx.Source(), id, 3)
				}), effects.AwakenLandTargets()),
				"Noyan Dar, Roil Shaper — put three +1/+1 counters on target land you control?"),
		},
	})
}
