package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viridescent Wisps — Instant {G} (Shadowmoor, #5591):
//
//	"Target creature becomes green and gets +1/+0 until end of turn.
//	 Draw a card."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3003e19e-ae60-4e10-b688-d0cac2a63cb2",
		Name:         "Viridescent Wisps",
		Purpose:      viridescentWispsPurpose(),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.OldB537WispsResolve("Viridescent Wisps", "G", "green",
			func(ctx *effects.Context, t game.TargetRef) error {
				return effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Viridescent Wisps — +1/+0"}.Apply(ctx)
			}),
	})
}

func viridescentWispsPurpose() game.Purpose {
	p := effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0))
	p.Draws = 1
	return p
}
