package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coral Helm — Artifact {3}:
//
//	"{3}, Discard a card at random: Target creature gets +2/+2 until end
//	 of turn."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa2970c8-f2ea-4e06-8b8f-ec89af0012a0",
		Name:         "Coral Helm",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, Discard a card at random: Target creature gets +2/+2 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)), game.AnswerPump),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.DiscardAtRandom(1, "a card at random")),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Coral Helm — +2/+2"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
