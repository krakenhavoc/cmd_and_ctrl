package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Light from Within — {2}{W}{W} Enchantment:
//
//	"Chroma — Each creature you control gets +1/+1 for each white mana symbol
//	 in its mana cost."
//
// Each creature reads its own printed cost, so a face-down creature (no mana
// cost, CR 708.2a) and a token get nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d6ad734-b8b5-4d91-9b7b-1fa0c55bfa4f",
		Name:         "Light from Within",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Chroma — Each creature you control gets +1/+1 for each white mana symbol in its mana cost.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, target *game.Card, _ *game.Game, source *game.Card) {
				n := effects.S82OldB521Chroma(target, source.TextColor("W"))
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
