package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mabel's Mettle — Instant {1}{W} (Bloomburrow, #4467):
//
//	"Target creature gets +2/+2 until end of turn. Up to one other target
//	 creature gets +1/+1 until end of turn."
//
// Two target clauses (CR 115.3: "other" is Distinct, so the second pick
// can't be the first): slot 0 gets +2/+2, the optional slot 1 +1/+1.
// Each slot is re-checked on its own as the spell resolves (CR 608.2b),
// so a creature that left in response is skipped and the other still
// gets its pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c657626f-7444-4c26-842c-dda9b9656fa3",
		Name:         "Mabel's Mettle",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.TargetCreature("up to one other target creature").WithCount(0, 1)),
		),
		Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2), effects.BenefitUntilEOT(1, 1, 1)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Mabel's Mettle — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				return effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Mabel's Mettle — +1/+1"}.Apply(ctx)
			}
			return nil
		},
	})
}
