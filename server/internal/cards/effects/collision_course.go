package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "49ca0ca8-0733-4aa9-bd0c-1c661725b81b",
		Name:         "Collision Course",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Collision Course deals X damage to target creature, where X is the number of permanents you control that are creatures and/or Vehicles.",
				TargetCreature("target creature"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					n := creaturesAndVehiclesControlledBy(ctx.Game, ctx.Controller())
					if n <= 0 {
						return nil
					}
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
				}),
			ModeDoing("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact()),
				DestroyTheModesTarget),
		),
	})
}
