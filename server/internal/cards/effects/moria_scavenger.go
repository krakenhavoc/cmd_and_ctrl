package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "a598e824-ad8e-4947-896b-70c5275b615a",
		Name:            "Moria Scavenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Discard a card: Draw a card. If the discarded card was a creature card, amass Orcs 1.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerMakesBlocker},
			Cost:    Plus(TapCost(), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				if card, ok := ctx.DiscardedCard(); ok && card.IsCreature() {
					return Amass{Subtype: "Orc", N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
