package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Cadets — Creature — Goblin {R}, 2/1 (Commander 2011, #5574):
//
//	"Whenever this creature blocks or becomes blocked, target opponent gains
//	 control of it. (This removes this creature from combat.)"
//
// One trigger over both events (Bushido's ThisBlocksOrBecomesBlocked), aimed at
// an opponent as it goes on the stack. The chosen opponent gains control for
// good (CR 611.2a, an indefinite layer-2 effect), which takes the Cadets out of
// combat (CR 506.4). If the Cadets have left or the opponent is gone by
// resolution nothing changes hands (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32dc847a-d8e9-4bc5-aa73-dfe69ff330a2",
		Name:         "Goblin Cadets",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OnAny([]game.EventKind{game.EventBlock, game.EventBecomesBlocked}, effects.ThisBlocksOrBecomesBlocked,
				"Goblin Cadets — target opponent gains control of it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					opp, ok := ctx.ClauseTarget(0)
					if !ok || opp.Kind != game.TargetPlayer || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.GainControl{
						Target:     item.SourceCardID,
						Controller: opp.ID,
						Duration:   game.IndefiniteDuration(),
						Label:      "Goblin Cadets — target opponent gains control of it",
					}.Apply(ctx)
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
