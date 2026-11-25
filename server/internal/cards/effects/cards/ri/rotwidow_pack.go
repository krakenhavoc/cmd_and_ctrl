package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rotwidow Pack — Creature — Spider {2}{B}{G}, 2/4 (Modern Horizons):
//
//	"Reach
//	 {3}{B}{G}, Exile a creature card from your graveyard: Create a 1/2
//	 green Spider creature token with reach, then each opponent loses 1 life
//	 for each Spider you control."
//
// Reach is a printed keyword. Exiling the creature card is part of the
// cost (CR 118.3), so the ability cannot be activated with none in the
// graveyard. On resolution the token is created first, then the Spiders
// you control are counted through the layers (the Pack and the token
// included) and every opponent loses that much life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "27b5d9d8-cc19-46be-9862-405cf7ef5408",
		Name:            "Rotwidow Pack",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}{G}, Exile a creature card from your graveyard: Create a 1/2 green Spider creature token with reach, then each opponent loses 1 life for each Spider you control.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{B}{G}"), effects.ExileFromGraveyard(1, "a creature card", effects.MatchCreature)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.CreateToken{Template: effects.TokenCard("1/2 green Spider with reach"), N: 1}).Apply(ctx); err != nil {
					return err
				}
				n := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtypeWord("Spider") })
				if n <= 0 {
					return nil
				}
				return effects.EachOpponentLosesLifeNow(g, item, n)
			},
		}},
	})
}
