package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Consuming Vortex — Instant — Arcane {1}{U}:
//
//	"Return target creature to its owner's hand.
//	Splice onto Arcane {3}{U}"
func init() {
	effects.Register(effects.Spec{
		OracleID:     "863bb8d8-4571-4775-87e3-d6f3ec88c835",
		Name:         "Consuming Vortex",
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{3}{U}"),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.BounceToHand{Target: t.ID}.Apply(ctx)
		},
	})
}
