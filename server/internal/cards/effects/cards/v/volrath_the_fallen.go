package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volrath the Fallen — Legendary Creature — Phyrexian Shapeshifter
// {3}{B}{B}{B}, 6/4:
//
//	"{1}{B}, Discard a creature card: Volrath gets +X/+X until end of
//	 turn, where X is the discarded card's mana value."
//
// ADR 0109 §8 (#1862): X is the discarded card's mana value, read off
// the payment record (Context.DiscardedManaValue, CR 400.7j) as the
// ability resolves and fixed then (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e11c838-68a7-4c97-b064-a46833cf1d2f",
		Name:         "Volrath the Fallen",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Discard a creature card: Volrath gets +X/+X until end of turn, where X is the discarded card's mana value.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.DiscardCardsMatching(1, "a creature card", effects.MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.DiscardedManaValue()
				return effects.BoostUntilEOT{
					Target:    item.SourceCardID,
					Power:     x,
					Toughness: x,
					Label:     "Volrath the Fallen — +X/+X",
				}.Apply(ctx)
			},
		}},
	})
}
