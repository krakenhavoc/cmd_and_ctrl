package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cosmogoyf — "Cosmogoyf's power is equal to the number of cards you
// own in exile and its toughness is equal to that number plus 1."
//
// A layer 7a characteristic-defining ability (CR 604.3, 613.4a), the
// Tarmogoyf shape: the count is every card in the exile zone whose owner
// is the Goyf's controller. The exile zone holds cards only (a token
// that is exiled ceases to exist), so no further filter is needed.
func init() {
	Register(Spec{
		OracleID:     "99d8c5f9-ad8f-450e-9a7c-51a4c0ca5043",
		Name:         "Cosmogoyf",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:    game.Layer7PT,
				Label:    "Power is the number of cards you own in exile; toughness is that plus 1.",
				SubLayer: game.SubLayer7A_CDA,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, source *game.Card) {
					n := 0
					if g != nil && g.Exile != nil {
						for i := range g.Exile.Cards {
							if g.Exile.Cards[i].Owner == source.Controller {
								n++
							}
						}
					}
					c.Power = n
					c.Toughness = n + 1
				},
			},
		},
	})
}
