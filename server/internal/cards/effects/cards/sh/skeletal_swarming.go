package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skeletal Swarming — Enchantment {3}{B}{G} (Adventures in the Forgotten
// Realms, S82):
//
//	"Each Skeleton you control has trample, attacks each combat if able, and
//	 gets +X/+0, where X is the number of other Skeletons you control.
//	 At the beginning of your end step, create a tapped 1/1 black Skeleton
//	 creature token. If a creature died this turn, create two of those tokens
//	 instead."
//
// Three statics over the Skeletons you control: the tribal keyword grant,
// the each-combat attack requirement and a +X/+0 worked out per Skeleton.
// The end-step trigger has no intervening "if": the death is read as it
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93ab7462-ceef-4bc7-a9fe-8af8ceb78e1c",
		Name:         "Skeletal Swarming",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.S82AfrB07SkeletonFilter, "trample"),
			effects.AttacksEachCombatWhere(effects.S82AfrB07SkeletonFilter.Matches),
			effects.S82AfrB07SkeletonPump(),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Skeletal Swarming — create a tapped 1/1 black Skeleton token, or two if a creature died this turn",
				func(g *game.Game, item *game.StackItem) error {
					n := 1
					if g.TurnTally.CreaturesDied > 0 {
						n = 2
					}
					return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("1/1 black Skeleton")).EntersTapped(), N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
