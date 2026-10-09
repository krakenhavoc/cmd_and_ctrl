package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ghalta the Unstoppable — Legendary Creature — Elder Dinosaur {8}{G}, 8/8:
//
//	"This spell costs {X} less to cast, where X is the greatest power
//	 among creatures you control.
//	 Trample
//	 Other creatures you control have trample."
//
// Ghalta, Primal Hunger's self cost modifier with the greatest power
// in place of the total (greatestPowerAmongCreaturesYouControl, the Great Henge's X). The
// reduction spends generic mana only, and a board with no creature
// reduces nothing. The trample grant is a layer 6 static on the OTHER
// creatures; Ghalta's own trample is its printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8e5ea773-a60b-4502-9db5-ce93fa61cc89",
		Name:            "Ghalta the Unstoppable",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(greatestPowerAmongCreaturesYouControl(),
				"This spell costs {X} less to cast, where X is the greatest power among creatures you control."),
		},
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID &&
					target.Controller == source.Controller && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, a := range c.Abilities {
					if a == "trample" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "trample")
			},
		}},
	})
}
