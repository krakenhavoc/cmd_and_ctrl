package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tapestry Warden — Artifact Creature — Robot Soldier {3}{G}, 3/4 (Edge of
// Eternities, #4461):
//
//	"Vigilance
//	 Each creature you control with toughness greater than its power
//	 assigns combat damage equal to its toughness rather than its power.
//	 Each creature you control with toughness greater than its power
//	 stations permanents using its toughness rather than its power."
//
// The combat half is Ghalta the Immovable's: a layer-6 static sets
// Characteristic.DamageByToughness on every creature the controller has, and
// the combat damage pass answers with the toughness only when it is greater
// than the power, per creature at damage time. The Warden itself is a 3/4,
// so it qualifies. The station half is CR 702.184c's modifier, the seam
// Interface Ace's crewing built (CostPowerRule) with one more clause,
// UseToughnessIfGreater, so the toughness replaces the power only when it is
// larger at the moment the cost is paid and again as the ability resolves
// (CR 608.2h): no creature's power changes, and a creature that left, or a
// Warden that has, is counted as the rulings say. The rule covers every
// station ability, a Planet's as well as a Spacecraft's
// (game.CostPowerStationAny beside game.CostPowerStation).
//
// Seam: CostPowerRule.UseToughnessIfGreater and CostPowerStationAny.
func init() {
	yourCreatures := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature()
	}
	Register(Spec{
		OracleID:        "981138d0-bf61-406a-85be-68cc5e58796c",
		Name:            "Tapestry Warden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			Label:     "Each creature you control with toughness greater than its power assigns combat damage equal to its toughness rather than its power.",
			AppliesTo: yourCreatures,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughness = true
			},
		}, {
			Layer:     game.Layer6Ability,
			Label:     "Each creature you control with toughness greater than its power stations permanents using its toughness rather than its power.",
			AppliesTo: yourCreatures,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.CostPower = append(c.CostPower, game.CostPowerRule{
					Uses:                  game.CostPowerStation | game.CostPowerStationAny,
					UseToughnessIfGreater: true,
				})
			},
		}},
	})
}
