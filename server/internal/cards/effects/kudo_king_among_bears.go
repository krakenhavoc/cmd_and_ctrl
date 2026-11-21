package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kudo, King Among Bears — Legendary Creature — Bear {G}{W}, 2/2 (Modern
// Horizons 3):
//
//	"Other creatures have base power and toughness 2/2 and are Bears in addition
//	 to their other types."
//
// Layer 4 adds the subtype, layer 7b sets the base size; counters and later
// effects still apply on top (CR 613.4).
func init() {
	Register(Spec{
		OracleID:     "070d22cf-2eed-4d04-b932-4b8339ea9b7a",
		Name:         "Kudo, King Among Bears",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Other creatures are Bears in addition to their other types.",
				Layer:     game.Layer4Type,
				AppliesTo: gapB52OtherCreatures,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !eotHasType(c.Subtypes, "Bear") {
						c.Subtypes = append(append([]string(nil), c.Subtypes...), "Bear")
					}
				},
			},
			{
				Label:     "Other creatures have base power and toughness 2/2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: gapB52OtherCreatures,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power = 2
					c.Toughness = 2
				},
			},
		},
	})
}
