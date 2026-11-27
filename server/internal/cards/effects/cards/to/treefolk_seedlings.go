package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treefolk Seedlings — Creature — Treefolk {2}{G}, 2/* (Urza's Saga, S82 OLD
// batch 731, #5624):
//
//	"Treefolk Seedlings's toughness is equal to the number of Forests you
//	 control."
//
// A characteristic-defining ability (CR 604.3, 613.4a): the toughness is the
// count of Forests its controller controls, layer 7a, so it moves as lands come
// and go and a Forest that changes hands changes it. "Forest" is the land
// subtype, read through the layers (a land that has become a Forest counts,
// one that has lost the type does not). With no Forest it is 2/0 and dies as a
// state-based action (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87285539-29c8-4cc7-a22c-4a4b96a61e82",
		Name:         "Treefolk Seedlings",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Toughness is equal to the number of Forests you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Toughness = effects.CountPermanentsYouControl(g, source, func(o *game.Card) bool {
					return o.IsLand() && o.HasSubtypeWord("Forest")
				})
			},
		}},
	})
}
