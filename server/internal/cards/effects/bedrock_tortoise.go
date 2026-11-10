package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bedrock Tortoise — Creature — Turtle {3}{G}, 2/6 (The Lost Caverns of
// Ixalan, #4470):
//
//	"During your turn, creatures you control have hexproof.
//	 Each creature you control with toughness greater than its power
//	 assigns combat damage equal to its toughness rather than its power."
//
// Two layer-6 statics. The hexproof is a keyword grant that applies only
// while it is the controller's turn, read live (CR 611.3a), so a creature
// is targetable on an opponent's turn. The damage clause is Ghalta the
// Immovable's flag (Characteristic.DamageByToughness): it is set on every
// creature the controller has, and the combat damage pass uses the
// toughness only when that is greater than the power (CR 510.1c), so a
// 4/4 still assigns 4.
//
// No simplification.
func init() {
	hexproof := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature() && isActivePlayer(g, source.Controller)
	}, "hexproof")
	hexproof.Label = "During your turn, creatures you control have hexproof."
	Register(Spec{
		OracleID:     "d9c2e674-11a9-47f0-8790-a53b62c43fe6",
		Name:         "Bedrock Tortoise",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{hexproof, {
			Layer: game.Layer6Ability,
			Label: "Each creature you control with toughness greater than its power assigns combat damage equal to its toughness rather than its power.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughness = true
			},
		}},
	})
}
