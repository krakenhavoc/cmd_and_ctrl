package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wings of the Cosmos — Instant {W} (Kaldheim, #5505):
//
//	"Target creature gets +1/+3 and gains flying until end of turn.
//	 Untap it."
//
// Acrobatic Leap's text, line for line: the boost, the keyword, then
// the untap, all on the one target if it is still legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "afa9cf24-cd02-41f5-9141-f6ff6f27d528",
		Name:         "Wings of the Cosmos",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3, "flying")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 3, Label: "Wings of the Cosmos — +1/+3"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Wings of the Cosmos — flying"}).Apply(ctx); err != nil {
					return err
				}
				return effects.UntapTarget{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
