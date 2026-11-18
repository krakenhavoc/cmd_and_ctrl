package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "bd8ba76f-d9e1-4c4d-83e1-6782fa3d01ba",
		Name:         "Timberland Guide",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Timberland Guide — put a +1/+1 counter on target creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target, ok := ctx.ClauseTarget(0)
					if !ok || target.Kind != game.TargetCard {
						return nil
					}
					return AddCounter{Target: target.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}), TargetCreature("target creature")),
		},
	})
}
