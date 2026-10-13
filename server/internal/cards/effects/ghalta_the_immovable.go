package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghalta the Immovable — Legendary Creature — Elder Dinosaur {8}{W}, 0/7:
//
//	"This spell costs {X} less to cast, where X is the greatest
//	 toughness among creatures you control.
//	 Creatures you control can attack as though they didn't have
//	 defender.
//	 Each creature you control with toughness greater than its power
//	 assigns combat damage equal to its toughness rather than its
//	 power."
//
// The cost reduction is Ghalta, Primal Hunger's self modifier keyed to
// toughness. "Can attack as though they didn't have defender" is
// written the way Stalked Researcher writes it, as the creatures
// losing the keyword while the static applies (a layer 6 removal on
// every creature its controller has), which is observably the same:
// defender only ever stops a creature attacking, and a creature with
// it still blocks.
//
// "Assigns combat damage equal to its toughness" is
// Characteristic.DamageByToughness: a layer 6 static sets the flag on
// every creature its controller has, and the combat damage pass
// (combatDamagePower, once per combatant per step) answers with the
// toughness whenever that is greater than the power (CR 510.1c). The
// check is per creature at damage time, so a 2/5 Wall assigns 5 and a
// 4/4 still assigns 4. Trample, deathtouch, lifelink and commander
// damage all read that number. No simplification. Doran, the Siege
// Tower and Assault Formation are not in the catalog yet and would
// reuse the flag.
func init() {
	Register(Spec{
		OracleID:     "0d496c6e-8f7b-420f-9197-dacd2feb528f",
		Name:         "Ghalta the Immovable",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(rfCreatureBGreatestToughnessYouControl(),
				"This spell costs {X} less to cast, where X is the greatest toughness among creatures you control."),
		},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Creatures you control can attack as though they didn't have defender.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				kept := c.Abilities[:0:0]
				for _, a := range c.Abilities {
					if a != "defender" {
						kept = append(kept, a)
					}
				}
				c.Abilities = kept
			},
		}, {
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
