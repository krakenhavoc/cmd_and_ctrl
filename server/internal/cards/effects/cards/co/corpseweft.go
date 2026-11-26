package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corpseweft — Enchantment {2}{B} (Dragons of Tarkir):
//
//	"{1}{B}, Exile one or more creature cards from your graveyard: Create a
//	 tapped X/X black Zombie Horror creature token, where X is twice the number
//	 of cards exiled this way."
//
// The cost names at least one creature card (game.ExileCost with MinCards 1); X
// is read from the cards the cost really exiled (Context.Exiled, CR 601.2h).
// The token's size is set as it is created, which a fixed template can't be, so
// its row is a 0/0 in knownUnmatchedTokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ffb5615-124e-4848-9e0d-120851bd50e5",
		Name:         "Corpseweft",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{B}, Exile one or more creature cards from your graveyard: Create a tapped X/X black Zombie Horror creature token, where X is twice the number of cards exiled this way.",
			// The token enters tapped, so it cannot block this turn: nothing to answer.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: effects.Plus(effects.ManaCost("{1}{B}"),
				effects.ExileOneOrMoreFromGraveyard("one or more creature cards", func(c game.Card) bool { return c.IsCreature() })),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := 2 * len(ctx.Exiled())
				tok := ctx.TextToken(effects.TokenCard("0/0 black Zombie Horror"))
				tok.Power, tok.Toughness = x, x
				tok.Tapped = true
				return ctx.Game.CreateTokenForEffect(ctx.Controller(), tok, 1)
			},
		}},
	})
}
