package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Baton of Courage — Artifact {3}:
//
//	"Flash
//	 Sunburst (This artifact enters with a charge counter on it for
//	 each color of mana spent to cast it.)
//	 Remove a charge counter from this artifact: Target creature gets
//	 +1/+1 until end of turn."
//
// Sunburst is a keyword the engine reads off the resolving spell (ADR
// 0109 §11, #1552); flash is printed.
func init() {
	effects.Register(effects.Spec{
		OracleID:            "a20bb8b3-b724-4fec-b179-ca9e6adc4718",
		Name:                "Baton of Courage",
		ReadsManaSpent:      game.ReadsOwnPayment,
		Completeness:        effects.CompletenessFull,
		WantsDistinctColors: true,
		PrintedKeywords:     []string{"flash", game.KeywordSunburst},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove a charge counter from this artifact: Target creature gets +1/+1 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)), game.AnswerPump),
			Cost:    effects.RemoveCountersFromThis(game.CounterCharge, 1),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
