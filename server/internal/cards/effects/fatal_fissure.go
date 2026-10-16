package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fatal Fissure — Instant {1}{B}:
//
//	"Choose target creature. When that creature dies this turn, you
//	 earthbend 4. (Target land you control becomes a 0/0 creature with
//	 haste that's still a land. Put four +1/+1 counters on it. When it
//	 dies or is exiled, return it to the battlefield tapped.)"
//
// The spell only creates a delayed triggered ability about the chosen
// creature as the object it is now (CR 603.7a, 603.7c; #2057), controlled
// by the caster (CR 603.7d), whoever controls the creature. A creature
// that leaves and comes back is a new object and does not fire it. The
// earthbend's "target land you control" is chosen as the trigger goes on
// the stack (CR 603.3d); with no land to choose, the trigger is removed.
// The earthbend itself is game.EarthbendForEffect (#1178).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "202e45d6-8c7a-46db-85fb-634aa77b4097",
		Name:         "Fatal Fissure",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			creature, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return whenThatCreatureDiesThisTurn(ctx, creature, "Fatal Fissure — earthbend 4", fatalFissureDiesBody)
		},
	})
}
