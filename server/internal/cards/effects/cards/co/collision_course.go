package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Collision Course — Sorcery {1}{W} (Aetherdrift, #4464):
//
//	"Choose one —
//	 • Collision Course deals X damage to target creature, where X is the
//	   number of permanents you control that are creatures and/or
//	   Vehicles.
//	 • Destroy target artifact."
//
// Crash and Burn's modal shape. The first bullet's X is Born to Drive's
// count (creaturesAndVehiclesControlledBy), each permanent once, so a
// crewed Vehicle counts one, read as the bullet resolves (CR 608.2h). The
// target may be any creature, yours included. The second bullet destroys
// any artifact, a Vehicle among them. Each bullet's target is re-checked
// as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49ca0ca8-0733-4aa9-bd0c-1c661725b81b",
		Name:         "Collision Course",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Collision Course deals X damage to target creature, where X is the number of permanents you control that are creatures and/or Vehicles.",
				effects.TargetCreature("target creature"),
				func(item *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					n := effects.CreaturesAndVehiclesControlledBy(ctx.Game, ctx.Controller())
					if n <= 0 {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
				}),
			effects.ModeDoing("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				effects.DestroyTheModesTarget),
		),
	})
}
