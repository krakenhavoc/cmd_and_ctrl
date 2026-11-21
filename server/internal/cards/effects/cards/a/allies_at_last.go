package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Allies at Last — Instant {2}{G} (Avatar: The Last Airbender, #4459):
//
//	"Affinity for Allies (This spell costs {1} less to cast for each Ally
//	 you control.)
//	 Up to two target creatures you control each deal damage equal to
//	 their power to target creature an opponent controls."
//
// Affinity is the shared self cost modifier, which reduces only the
// generic part of the cost (the first ruling). Two target clauses, as
// Terrific Team-Up prints them: slot 0 up to two creatures you control
// (the damage sources), slot 1 the recipient. On resolution each source
// still legal deals damage equal to its current power to the recipient
// as one damage event; a source that has left deals nothing, and a
// recipient that has left takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c15e5e93-2eff-407d-ae75-437e36ef85ba",
		Name:         "Allies at Last",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("up to two target creatures you control", effects.YouControl()).WithCount(0, 2),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Allies", effects.HasSubtype("Ally")),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var sources []game.TargetRef
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					sources = append(sources, t)
				}
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			return effects.EachDealsDamageEqualToItsPower(ctx, sources, victim.ID)
		},
	})
}
