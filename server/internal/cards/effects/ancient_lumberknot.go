package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancient Lumberknot — {2}{B}{G} Creature — Treefolk 1/4 (Innistrad:
// Crimson Vow, S82):
//
//	"Each creature you control with toughness greater than its power
//	 assigns combat damage equal to its toughness rather than its
//	 power."
//
// The same layer 6 flag as Ghalta the Immovable
// (Characteristic.DamageByToughness, CR 510.1c), set on every creature
// its controller has, Lumberknot included. No simplification.
func init() {
	Register(Spec{
		OracleID:     "d714f1f6-539e-4a2a-b738-226854d7ff09",
		Name:         "Ancient Lumberknot",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
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
