package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coiling Woodworm — Creature — Insect Worm {2}{G}, */1 (Nemesis, S82 OLD
// batch 680, #5618):
//
//	"Coiling Woodworm's power is equal to the number of Forests on the
//	 battlefield."
//
// A characteristic-defining ability (CR 604.3, layer 7a) counted live over
// every Forest, any player's, a land made a Forest included (CR 205.3i).
// The Forest word is the source's text (CR 612.2). Toughness stays the
// printed 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75dcdf05-b392-4a96-98ae-5ec4cd425eee",
		Name:         "Coiling Woodworm",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			Label:     "Coiling Woodworm's power is equal to the number of Forests on the battlefield.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				forest := source.TextSubtype("Forest")
				n := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.HasSubtype(forest) {
						n++
					}
				}
				c.Power = n
			},
		}},
	})
}
