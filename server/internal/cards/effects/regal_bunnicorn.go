package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Regal Bunnicorn — Creature — Rabbit Unicorn {1}{W}, */* (Wilds of
// Eldraine, #4471):
//
//	"Regal Bunnicorn's power and toughness are each equal to the number of
//	 nonland permanents you control."
//
// Burrowguard Mentor's layer 7a characteristic-defining ability (CR
// 604.3, 613.4a) on both numbers: self-only, set on every recompute to
// the count of nonland permanents its controller controls, the Bunnicorn
// itself included. A land animated into a creature is still a land and
// does not count; an artifact, enchantment, planeswalker or token does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "efd93428-88a5-4f7b-95f3-1d6eec069a4c",
		Name:         "Regal Bunnicorn",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Regal Bunnicorn's power and toughness are each equal to the number of nonland permanents you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for i := range g.Battlefield.Cards {
					if p := &g.Battlefield.Cards[i]; p.Controller == source.Controller && !p.IsLand() {
						n++
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
