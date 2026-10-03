package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Living Plane — World Enchantment {2}{G}{G}:
//
//	"All lands are 1/1 creatures that are still lands."
//
// March of the World Ooze's two layers over every land on the
// battlefield: layer 4 adds the creature type, keeping every other
// type and subtype, and layer 7b sets the base power and toughness to
// 1/1, so +1/+1 counters and pumps still apply on top. The lands are
// summoning sick the turn they enter, as any creature is (CR 302.6),
// which includes their {T} mana abilities.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30e99293-3212-4e3f-b543-b1c7c416575d",
		Name:         "Living Plane",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer4Type,
				AppliesTo: allLands,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, t := range c.Types {
						if equalFoldASCIIEffects(t, "Creature") {
							return
						}
					}
					c.Types = append(c.Types, "Creature")
				},
			},
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: allLands,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 1, 1
				},
			},
		},
	})
}

// allLands is the static scope "all lands" — every land on the
// battlefield, every player's.
func allLands(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.IsLand()
}
