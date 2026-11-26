package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Springjack Shepherd — {3}{W} Creature — Kithkin Wizard, 1/2:
//
//	"Chroma — When this creature enters, create a 0/1 white Goat creature token
//	 for each white mana symbol in the mana costs of permanents you control."
//
// Read as the trigger resolves, so the Shepherd's own {W} counts.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "325b3469-531e-4ca2-bccb-bb95258c1b54",
		Name:         "Springjack Shepherd",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Springjack Shepherd — a Goat for each white mana symbol", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := effects.S82OldB521ChromaOfPermanents(g, item.Controller, ctx.TextColor("W"))
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("0/1 white Goat"), N: n}.Apply(ctx)
			}),
		},
	})
}
