package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kumena's Speaker — Creature — Merfolk Shaman {G}, 1/1 (Ixalan):
//
//	"This creature gets +1/+1 as long as you control another Merfolk or an
//	 Island."
//
// One layer 7c static on the source alone, live through the layers: another
// Merfolk (the Speaker never counts itself) or any Island its controller
// controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68b26aeb-4952-4ca1-9b4a-82bdca16461b",
		Name:         "Kumena's Speaker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control another Merfolk or an Island.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.OldB221SelfWhileAnother([]string{"Merfolk"}, []string{"Island"}),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power++
				c.Toughness++
			},
		}},
	})
}
