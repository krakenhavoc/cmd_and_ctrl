package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// S82OldB195LandsPower is a characteristic-defining ability that sets only
// the creature's power to the number of lands its controller controls
// (CR 604.3, 613.4a). Zendikar Incarnate's toughness is printed.
func S82OldB195LandsPower(label string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		Label:     label,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			c.Power = B10LandsControlled(g, source.Controller)
		},
	}
}
