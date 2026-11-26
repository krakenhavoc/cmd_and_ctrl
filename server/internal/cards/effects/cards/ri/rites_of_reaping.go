package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rites of Reaping — Sorcery {4}{B}{G} (Return to Ravnica, S82 OLD):
//
//	"Target creature gets +3/+3 until end of turn. Another target creature
//	 gets -3/-3 until end of turn."
//
// Two target clauses, the second a different creature (CR 601.2c). Each
// instruction reads its own clause (CR 608.2b), so one going away does not
// stop the other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed1a1c04-1b2b-4616-abb2-8bd19c47b589",
		Name:         "Rites of Reaping",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3), effects.HarmUntilEOT(1, -3, -3)),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Rites of Reaping — +3/+3 until end of turn"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.BoostUntilEOT{Target: t.ID, Power: -3, Toughness: -3, Label: "Rites of Reaping — -3/-3 until end of turn"}.Apply(ctx)
			}),
		),
	})
}
