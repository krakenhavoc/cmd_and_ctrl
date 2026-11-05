package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Traveler — Legendary Creature — Scout {G}{U}, */3 (Star Trek,
// #4973):
//
//	"Reach, vigilance
//	 Strange new worlds — The Traveler's power is equal to the number of
//	 differently named lands you control."
//
// Reach and vigilance are the printed keywords. "Strange new worlds" is
// an ability word with no rules meaning. The power is a layer 7a
// characteristic-defining ability (CR 604.3, 613.4a), Mendicant Core's
// shape, counting the distinct names among the lands its controller
// controls (socB23DifferentlyNamedLands's count), so two Forests are one
// name and a Forest and a Snow-Covered Forest are two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b61718eb-dc6a-4199-b7c7-42eccf4158e8",
		Name:            "The Traveler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Static: []game.StaticAbility{{
			Label:     "Strange new worlds — The Traveler's power is equal to the number of differently named lands you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				names := map[string]bool{}
				for _, p := range g.Battlefield.Cards {
					if p.Controller == source.Controller && p.IsLand() {
						names[p.Name] = true
					}
				}
				c.Power = len(names)
			},
		}},
	})
}
