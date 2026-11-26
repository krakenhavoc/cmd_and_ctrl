package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cruel Finality — Instant {2}{B} (Aether Revolt, S82 older sets, #5536):
//
//	"Target creature gets -2/-2 until end of turn. Scry 1."
//
// The scry is the second sentence of the same spell, so it happens even if
// the creature has died to the -2/-2 (CR 608.2c); a target that is gone
// before resolution fizzles the whole spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1285a519-b3a8-4014-8f33-89f3beb0c00b",
		Name:         "Cruel Finality",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.HarmUntilEOT(0, -2, -2)), Scry: 1},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -2, Toughness: -2, Label: "Cruel Finality — -2/-2"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
			}),
		),
	})
}
