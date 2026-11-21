package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flowstone Infusion — Instant {R} (Dominaria United, #5487):
//
//	"Target creature gets +2/-2 until end of turn."
//
// One layer 7c modification on the target. A creature whose toughness
// reaches 0 dies to the state-based check (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c75b92a-4fc1-4f26-95be-1b92f462c111",
		Name:         "Flowstone Infusion",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: -2, Label: "Flowstone Infusion — +2/-2"}.Apply(ctx)
		},
	})
}
