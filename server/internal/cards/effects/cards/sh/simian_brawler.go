package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Simian Brawler — Creature — Ape Warrior {3}{G}, 3/3 (Jumpstart 2022,
// #5948):
//
//	"Discard a land card: This creature gets +1/+1 until end of turn."
//
// Discarding a land card is the cost, paid as the ability is announced (CR
// 602.2b) and not refunded if the ability is countered. There is no tap, so
// it can be activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2497cfe2-a85c-47ea-af7a-9a745d84f8b4",
		Name:         "Simian Brawler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a land card: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.DiscardCardsMatching(1, "a land card", func(c game.Card) bool { return c.IsLand() }),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Toughness: 1, Label: "Simian Brawler — +1/+1"}.Apply(ctx)
			},
		}},
	})
}
