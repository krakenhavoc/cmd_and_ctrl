package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b536_helpers.go — S82 OLD batch 536 (#5591, Shadowmoor creatures).

// PowerAndToughnessEqualToPermanentsOfColorYouControl is "<this>'s power and
// toughness are each equal to the number of <color> permanents you control"
// (Horde of Boggarts, Kithkin Rabble, Faerie Swarm, Crowd of Cinders): a
// characteristic-defining ability (CR 604.3, CR 613.4a), counted live in layer
// 7a over the permanents the source's controller controls, itself included when
// it is that colour. The colour is a word of the card's text (CR 612.2).
func PowerAndToughnessEqualToPermanentsOfColorYouControl(name, color, colorWord string) game.StaticAbility {
	return game.StaticAbility{
		Label:     name + "'s power and toughness are each equal to the number of " + colorWord + " permanents you control.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := CountPermanentsYouControl(g, source, func(p *game.Card) bool { return p.HasColorWord(color) })
			c.Power, c.Toughness = n, n
		},
	}
}
