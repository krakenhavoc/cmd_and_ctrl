package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woodborn Behemoth — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "9bc458ac-5566-4978-aeaa-73b600d4ec29", Name: "Woodborn Behemoth", Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{Label: "As long as you control eight or more lands, this creature gets +4/+4.", Layer: game.Layer7PT, SubLayer: game.SubLayer7C_Modify, AppliesTo: woodbornBehemothEightLands, Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) { c.Power += 4; c.Toughness += 4 }},
			{Label: "As long as you control eight or more lands, this creature has trample.", Layer: game.Layer6Ability, AppliesTo: woodbornBehemothEightLands, Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "trample")
			}},
		},
	})
}

func woodbornBehemothEightLands(target *game.Card, g *game.Game, source *game.Card) bool {
	return source != nil && target.InstanceID == source.InstanceID && b10LandsControlled(g, source.Controller) >= 8
}
