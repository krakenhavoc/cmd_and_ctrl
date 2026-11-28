package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// An-Havva Constable — Creature — Human {1}{G}{G}, 2/1+* (Homelands, S82 OLD
// b826, #5635):
//
//	"An-Havva Constable's toughness is equal to 1 plus the number of green
//	 creatures on the battlefield."
//
// A characteristic-defining ability (CR 604.3, layer 7a) on toughness alone,
// read live: every green creature on the battlefield counts, whoever controls
// it, and the Constable itself (a green creature) is one of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7131cdb6-15d0-407e-a7e2-b2a0deef9363",
		Name:         "An-Havva Constable",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "An-Havva Constable's toughness is equal to 1 plus the number of green creatures on the battlefield.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				c.Toughness = 1 + effects.OldB826GreenCreatures(g)
			},
		}},
	})
}
