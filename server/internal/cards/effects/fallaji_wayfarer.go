package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fallaji Wayfarer — Creature — Human Scout {2}{G}, 2/4 (Dominaria United,
// S82 DMU batch 15):
//
//	"Fallaji Wayfarer is all colors. This ability doesn't affect its color
//	 identity. (It can be in any deck whose commander's color identity
//	 includes green.)
//	 Multicolored spells you cast have convoke."
//
// The first line is a layer 5 set of all five colors on this permanent only
// (CR 105.2c, 613.1e); color identity is read from the printed card, so it
// stays green. The second is a convoke grant to every multicolored spell
// its controller casts, read against the spell's colors as it is cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "01de1f9d-3b63-49b9-b2c6-929c2a69ef54",
		Name:         "Fallaji Wayfarer",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Fallaji Wayfarer is all colors.",
			Layer: game.Layer5Color,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Colors = []string{"W", "U", "B", "R", "G"}
			},
		}},
		GrantedTapCosts: []game.GrantedTapCost{SpellsYouCastHaveConvoke(
			"Multicolored spells you cast have convoke.", SpellIs(Multicolored()))},
	})
}
