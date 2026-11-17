package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Multani, Maro-Sorcerer — Legendary Creature — Elemental Sorcerer
// {4}{G}{G}, */* (Urza's Legacy, #5623):
//
//	"Shroud (This creature can't be the target of spells or abilities.)
//	 Multani's power and toughness are each equal to the total number
//	 of cards in all players' hands."
//
// Shroud is a printed keyword (CR 702.18a). Its power and toughness are a
// characteristic-defining ability (CR 604.3) applied in layer 7a
// (CR 613.4a), Lord of Extinction's shape counting every player's hand
// instead of every graveyard, so it grows and shrinks as cards are drawn,
// played and discarded, everywhere it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bce5b327-7da1-4136-b76f-069e2fef5d2a",
		Name:            "Multani, Maro-Sorcerer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Static: []game.StaticAbility{{
			Label:    "Multani's power and toughness are each equal to the total number of cards in all players' hands.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := cardsInAllHands(g)
				c.Power = n
				c.Toughness = n
			},
		}},
	})
}
