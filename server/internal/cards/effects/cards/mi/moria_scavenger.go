package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moria Scavenger — Creature — Orc Rogue {1}{B}{R}, 1/4:
//
//	"Deathtouch, haste
//	 {T}, Discard a card: Draw a card. If the discarded card was a
//	 creature card, amass Orcs 1."
//
// ADR 0109 §8 (#1862): "the discarded card" is read off the payment
// record (Context.DiscardedCard, CR 400.7j) after the draw, in the
// printed order. The amass is the keyword action every amass card
// shares (Amass, CR 701.47).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a598e824-ad8e-4947-896b-70c5275b615a",
		Name:            "Moria Scavenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Discard a card: Draw a card. If the discarded card was a creature card, amass Orcs 1.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.TapCost(), effects.DiscardACard()),
			Effect: effects.Do(
				// The Army waits for a prompt the draw asks (CR 608.2c, #4155).
				effects.DrawCards{N: 1},
				effects.Instruction(func(ctx *effects.Context) error {
					if card, ok := ctx.DiscardedCard(); ok && card.IsCreature() {
						return effects.Amass{Subtype: "Orc", N: 1}.Apply(ctx)
					}
					return nil
				}),
			),
		}},
	})
}
