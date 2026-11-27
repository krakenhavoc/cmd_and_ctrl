package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Allosaurus Rider — Creature — Elf Warrior {5}{G}{G}, 1+*/1+* (Coldsnap, #5597):
//
//	"You may exile two green cards from your hand rather than pay this spell's
//	 mana cost.
//	 Allosaurus Rider's power and toughness are each equal to 1 plus the number
//	 of lands you control."
//
// Soul Spike's two-card pitch (an alternative cost, CR 118.9; the Rider itself,
// already on the stack, can't be one of the two) and Beanstalk Giant's
// characteristic-defining ability (layer 7a) with the 1 added, read live.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "176508ba-c107-4694-8dd7-ea34fca3b299",
		Name:         "Allosaurus Rider",
		Completeness: effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Pitch("Exile two green cards from your hand", 0,
				effects.CardInYourHand("two green cards from your hand", effects.OfColor("G")).WithCount(2, 2),
				"two green cards"),
		},
		Static: []game.StaticAbility{{
			Label:     "Allosaurus Rider's power and toughness are each equal to 1 plus the number of lands you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 1 + effects.B10LandsControlled(g, source.Controller)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
