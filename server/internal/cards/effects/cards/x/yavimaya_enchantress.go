package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yavimaya Enchantress — {2}{G} Creature — Human Druid 2/2:
//
//	"This creature gets +1/+1 for each enchantment on the battlefield."
//
// Layer 7c (CR 613.4c), counting every enchantment on the battlefield
// whoever controls it, after layer 4 has typed the permanents.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "495fc654-7986-4d76-9d5c-bc1484e98de2",
		Name:         "Yavimaya Enchantress",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 for each enchantment on the battlefield.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := 0
				for _, p := range g.Battlefield.Cards {
					if p.IsEnchantment() {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
