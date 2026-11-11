package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timber Paladin — Artifact Creature — Knight {1}{G}, 1/1 (Wilds of
// Eldraine Commander, #5368):
//
//	"As long as this creature is enchanted by exactly one Aura, it has
//	 base power and toughness 3/3.
//	 As long as this creature is enchanted by exactly two Auras, it has
//	 base power and toughness 5/5 and vigilance.
//	 As long as this creature is enchanted by three or more Auras, it has
//	 base power and toughness 10/10, vigilance, and trample."
//
// Three tiers, each a layer 7b base-P/T set under a count of the Auras
// attached to the Paladin right now (cmdr2B19AurasAttachedTo; any
// player's, Roles included), plus the layer 6 keyword grants of the upper
// two. Exactly one, exactly two and three or more are disjoint, so only
// one tier applies at a time. Pumps and counters apply on top (layer 7c
// and 7d), as the "base" in the text says.
//
// No simplification.
func init() {
	tier := func(label string, match func(n int) bool, power, toughness int, keywords ...string) []game.StaticAbility {
		applies := func(target *game.Card, g *game.Game, source *game.Card) bool {
			return selfOnly(target, g, source) && match(cmdr2B19AurasAttachedTo(g, source.InstanceID))
		}
		out := []game.StaticAbility{{
			Label:     label,
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: applies,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = power
				c.Toughness = toughness
			},
		}}
		if len(keywords) > 0 {
			out = append(out, game.StaticAbility{
				Label:     label,
				Layer:     game.Layer6Ability,
				AppliesTo: applies,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, kw := range keywords {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
					}
				},
			})
		}
		return out
	}
	var statics []game.StaticAbility
	statics = append(statics, tier("As long as this creature is enchanted by exactly one Aura, it has base power and toughness 3/3.",
		func(n int) bool { return n == 1 }, 3, 3)...)
	statics = append(statics, tier("As long as this creature is enchanted by exactly two Auras, it has base power and toughness 5/5 and vigilance.",
		func(n int) bool { return n == 2 }, 5, 5, "vigilance")...)
	statics = append(statics, tier("As long as this creature is enchanted by three or more Auras, it has base power and toughness 10/10, vigilance, and trample.",
		func(n int) bool { return n >= 3 }, 10, 10, "vigilance", "trample")...)
	Register(Spec{
		OracleID:     "0be44c76-9e3f-45c2-9b2c-6a5f00678780",
		Name:         "Timber Paladin",
		Completeness: CompletenessFull,
		Static:       statics,
	})
}
