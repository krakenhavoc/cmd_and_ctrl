package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savor — Instant {1}{B} (Bloomburrow, #4467):
//
//	"Target creature gets -2/-2 until end of turn. Create a Food token.
//	 (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3
//	 life.")"
//
// Puny Snack's shrink, then the catalog's printed Food token for the
// caster, as two steps (CR 608.2c). A target that left in response
// counters the spell (CR 608.2b), and then no Food is made either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfc1dc36-0cc2-47cc-b0dd-1cc7c31057aa",
		Name:         "Savor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Tokens: 1, Targets: game.ForTargets(effects.HarmUntilEOT(0, -2, -2))},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Savor — -2/-2"}.Apply(ctx)
			}),
			effects.CreateToken{Template: effects.FoodToken(), N: 1},
		),
	})
}
