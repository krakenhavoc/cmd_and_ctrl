package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Necromancer's Stockpile — Enchantment {1}{B}:
//
//	"{1}{B}, Discard a creature card: Draw a card. If the discarded
//	 card was a Zombie card, create a tapped 2/2 black Zombie creature
//	 token."
//
// ADR 0109 §8 (#1862): "the discarded card" is read off the payment
// record (Context.DiscardedCard, CR 400.7j) after the draw, in the
// printed order. "Zombie card" is the card's own creature type, as it
// is in the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc47a97b-f511-4b97-97b2-e0309055544b",
		Name:         "Necromancer's Stockpile",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Discard a creature card: Draw a card. If the discarded card was a Zombie card, create a tapped 2/2 black Zombie creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{B}"), DiscardCardsMatching(1, "a creature card", MatchCreature)),
			// The token waits for the draw's prompt (a dredge offer; CR 608.2c, #4155).
			Effect: Do(
				DrawCards{N: 1},
				Instruction(func(ctx *Context) error {
					card, ok := ctx.DiscardedCard()
					// "Zombie" is a word of the source's text (CR 612.2, #3984).
					if !ok || !card.HasSubtype(ctx.TextSubtype("Zombie")) {
						return nil
					}
					tmpl := TokenCard("2/2 black Zombie")
					tmpl.Tapped = true
					return CreateToken{Template: tmpl, N: 1}.Apply(ctx)
				}),
			),
		}},
	})
}
