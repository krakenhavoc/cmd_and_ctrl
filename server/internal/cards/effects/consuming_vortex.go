package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Consuming Vortex — Instant — Arcane {1}{U}:
//
//	"Return target creature to its owner's hand.
//	Splice onto Arcane {3}{U}"
func init() {
	Register(Spec{
		OracleID:     "863bb8d8-4571-4775-87e3-d6f3ec88c835",
		Name:         "Consuming Vortex",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{3}{U}"),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return BounceToHand{Target: t.ID}.Apply(ctx)
		},
	})
}
