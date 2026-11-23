package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Timberland Guide — Creature — Human Scout {1}{G}, 1/1 (Innistrad:
// Midnight Hunt Commander reprint, S82):
//
//	"When this creature enters, put a +1/+1 counter on target creature."
//
// Caldaia Strongarm's shape with one counter. The target is chosen as the
// trigger goes on the stack and may be the Guide itself or any other
// player's creature; a target gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd8ba76f-d9e1-4c4d-83e1-6782fa3d01ba",
		Name:         "Timberland Guide",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Timberland Guide — put a +1/+1 counter on target creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target, ok := ctx.ClauseTarget(0)
					if !ok || target.Kind != game.TargetCard {
						return nil
					}
					return effects.AddCounter{Target: target.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}), effects.TargetCreature("target creature")),
		},
	})
}
