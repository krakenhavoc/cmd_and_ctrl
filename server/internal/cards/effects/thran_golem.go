package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thran Golem — Artifact Creature — Golem {5}, 3/3 (Dominaria Remastered):
//
//	"As long as this creature is enchanted, it gets +2/+2 and has flying,
//	 first strike, and trample."
//
// "Enchanted" is Timber Paladin's reading: an Aura, any player's, is
// attached to it right now (cmdr2B19AurasAttachedTo). The +2/+2 is layer
// 7c and the keywords layer 6, both gated on the same count.
//
// No simplification.
func init() {
	enchanted := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && cmdr2B19AurasAttachedTo(g, source.InstanceID) > 0
	}
	Register(Spec{
		OracleID:     "fe55f366-2996-4a4f-9b30-fc494cc80f87",
		Name:         "Thran Golem",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as this creature is enchanted, it gets +2/+2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: enchanted,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			{
				Label:     "As long as this creature is enchanted, it has flying, first strike, and trample.",
				Layer:     game.Layer6Ability,
				AppliesTo: enchanted,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, kw := range []string{"flying", "first strike", "trample"} {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
					}
				},
			},
		},
	})
}
