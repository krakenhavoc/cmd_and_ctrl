package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Traproot Kami — Creature — Spirit {G}, 0/* (Betrayers of Kamigawa,
// #5602):
//
//	"Defender; reach (This creature can block creatures with flying.)
//	 Traproot Kami's toughness is equal to the number of Forests on the
//	 battlefield."
//
// A characteristic-defining ability (CR 604.3), applied in layer 7a like
// Dungrove Elder's, that sets toughness only: power stays the printed 0.
// "On the battlefield" counts every player's Forests, by effective
// subtype, so a land made a Forest counts and the word reads through a
// text change.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce7d555a-9197-4241-94ee-a164a65099a5",
		Name:            "Traproot Kami",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		Static: []game.StaticAbility{{
			Label:     "Traproot Kami's toughness is equal to the number of Forests on the battlefield.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				forest := source.TextSubtype("Forest")
				n := 0
				for _, p := range g.Seats {
					n += effects.B08LandsWithSubtypeControlled(g, p.ID, forest)
				}
				c.Toughness = n
			},
		}},
	})
}
