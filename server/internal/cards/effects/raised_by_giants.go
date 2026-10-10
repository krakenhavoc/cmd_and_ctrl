package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raised by Giants — Legendary Enchantment — Background {5}{G} (EDHREC
// rank 6078):
//
//	"Commander creatures you own have base power and toughness 10/10
//	 and are Giants in addition to their other types."
//
// Two characteristic-changing statics from the Background, both on
// each commander creature its controller owns, under anyone's control:
// the Giant type in layer 4 and base 10/10 in layer 7b (CR 613.1d,
// 613.4b). Anthems, +1/+1 counters and "gets +X/+X" still apply on
// top, in 7c, as they do on Blade of the Oni's base 5/5. Nothing is
// granted, so a later "loses all abilities" leaves it a 10/10 Giant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1682cf24-17a3-49ad-8b6f-9b7f13ebf53c",
		Name:         "Raised by Giants",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer4Type,
				AppliesTo: commanderCreatureYouOwn,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !hasFold(c.Subtypes, "Giant") {
						c.Subtypes = append(c.Subtypes, "Giant")
					}
				},
			},
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: commanderCreatureYouOwn,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 10, 10
				},
			},
		},
	})
}
