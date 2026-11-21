package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nuka-Cola Vending Machine — Artifact {3} (EDHREC rank 1294):
//
//	"{1}, {T}: Create a Food token. (It's an artifact with "{2}, {T},
//	 Sacrifice this token: You gain 3 life.")
//	 Whenever you sacrifice a Food, create a tapped Treasure token.
//	 (It's an artifact with "{T}, Sacrifice this token: Add one mana
//	 of any color.")"
//
// The Food deck's Treasure engine. The activation is a CR 602
// ability with a mana-and-tap cost; the trigger watches
// EventSacrifice, which fires before the zone move so the Food is
// still on the battlefield to be recognised by subtype (a Gingerbrute
// or a Food-typed token from another card counts, as printed). The
// Treasure enters tapped through CreateTokenAdvanced, and the Amulet
// of Vigor in the same batch sees that entry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6cfb03e5-aca6-4fe2-a3f1-93e1f0cbf9e1",
		Name:         "Nuka-Cola Vending Machine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Create a Food token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Controller: item.Controller, Template: effects.FoodToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B11SacrificedAFood(ev, source, g)
			}, "Nuka-Cola Vending Machine — create a tapped Treasure", effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TreasureToken()).EntersTapped(), N: 1})),
		},
	})
}
