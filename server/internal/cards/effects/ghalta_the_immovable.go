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
// Gap, declared as a caveat: the combat damage step reads each
// creature's power (assignAndDealCombatDamageLocked) and has no seam
// for "assigns combat damage equal to its toughness", so a Wall or
// Ghalta himself deals damage by power. That is weaker than printed,
// never stronger. Doran, the Siege Tower and Assault Formation wait on
// the same seam; the card moves to Full when it lands.
func init() {
	Register(Spec{
		OracleID:     "0d496c6e-8f7b-420f-9197-dacd2feb528f",
		Name:         "Ghalta the Immovable",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Creatures that have more toughness than power still deal combat damage equal to their power, not their toughness.",
		},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(rfCreatureBGreatestToughnessYouControl(),
				"This spell costs {X} less to cast, where X is the greatest toughness among creatures you control."),
		},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
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
		}},
	})
}
