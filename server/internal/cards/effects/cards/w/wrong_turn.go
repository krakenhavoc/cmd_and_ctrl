package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wrong Turn — Instant {2}{U} (Commander Legends):
//
//	"Target opponent gains control of target creature. (If an attacking or
//	 blocking creature changes controllers, it's removed from combat.)"
//
// Harmless Offering's gift with any creature as the second target: the
// opponent gains control for good (CR 611.2a, an indefinite layer-2 effect)
// and can't refuse. With either target illegal on resolution nothing changes
// hands (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1cfbbcc-4aee-4400-b5a8-18713829ae72",
		Name:         "Wrong Turn",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPlayer("target opponent", effects.Opponent()),
			effects.TargetCreature("target creature"),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			opp, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			creature, ok := ctx.ClauseTarget(1)
			if !ok || creature.Kind != game.TargetCard {
				return nil
			}
			return effects.GainControl{
				Target:     creature.ID,
				Controller: opp.ID,
				Duration:   game.IndefiniteDuration(),
				Label:      "Wrong Turn — target opponent gains control of it",
			}.Apply(ctx)
		},
	})
}
