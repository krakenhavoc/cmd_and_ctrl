package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fountainport — Land (EDHREC rank 928):
//
//	"{T}: Add {C}.
//	 {2}, {T}, Sacrifice a token: Draw a card.
//	 {3}, {T}, Pay 1 life: Create a 1/1 blue Fish creature token.
//	 {4}, {T}: Create a Treasure token."
//
// The Bloomburrow utility land: a colorless source that turns spare
// mana into a token, a Treasure, or — with a token to feed it — a
// card. Three activated abilities on one land, each with its own
// cost shape, all through the ordinary CR 602 path: mana and life
// validated before anything is paid, the token sacrifice picked from
// the activator's own tokens (any token — a Treasure, a Fish, a
// creature), and the tap shared, so the land does one thing a turn.
// The Fish is a real 1/1 blue creature token and the Treasure is
// tokens.go's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "94e8b0a9-44a1-4dce-8d44-78681ae638a1",
		Name:          "Fountainport",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}, {T}, Sacrifice a token: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.B08SacrificeAToken()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{3}, {T}, Pay 1 life: Create a 1/1 blue Fish creature token.",
				Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.PayLife(1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 blue Fish"), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{4}, {T}: Create a Treasure token.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
