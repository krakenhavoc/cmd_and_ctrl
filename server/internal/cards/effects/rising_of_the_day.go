package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rising of the Day — Enchantment {2}{R} (EDHREC rank 715):
//
//	"Creatures you control have haste.
//	 Legendary creatures you control get +1/+0."
//
// Two statics, two layers: the haste grant is Layer 6 (the Lord of
// Atlantis shape, keyword appended to every creature you control), the
// pump is Layer 7c and reads the POST-layer supertypes, so a creature
// made legendary by a type-changing effect gets it too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "434a20f8-3f87-4004-9155-0f196fc2257e",
		Name:         "Rising of the Day",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !eotHasAbility(c.Abilities, "haste") {
						c.Abilities = append(c.Abilities, "haste")
					}
				},
			},
			{
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller && b06IsLegendary(target)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
				},
			},
		},
	})
}
