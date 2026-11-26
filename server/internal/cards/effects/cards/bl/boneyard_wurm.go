package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boneyard Wurm — Creature — Wurm {1}{G}, */* (Duel Decks: Izzet vs. Golgari):
//
//	"Boneyard Wurm's power and toughness are each equal to the number of
//	 creature cards in your graveyard."
//
// Splinterfright's layer 7a characteristic-defining ability over the
// controller's own graveyard (CR 604.3, CR 613.4a): read live, so a creature
// card arriving in or leaving the graveyard moves it at once, and it dies as a
// state-based action at 0 toughness (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6e5f908a-f4f1-4718-ac39-34176e401756",
		Name:         "Boneyard Wurm",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Boneyard Wurm's power and toughness are each equal to the number of creature cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.B11CreatureCardsInGraveyard(g, source.Controller)
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
