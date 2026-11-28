package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glacial Chasm — Land (Ice Age, S82 OLD batch 838, #5636):
//
//	"Cumulative upkeep—Pay 2 life.
//	 When this land enters, sacrifice a land.
//	 Creatures you control can't attack.
//	 Prevent all damage that would be dealt to you."
//
// Cumulative upkeep with a life payment (LifePayment, multiplied by the age
// counters). The enters trigger is Foul Spirit's: its controller chooses the
// land as it resolves, the Chasm itself included (CR 701.21). "Creatures you
// control can't attack" is Evil Eye of Orms-by-Gore's layer 6 restriction over
// the controller's creatures, and the damage to you is prevented as Solitary
// Confinement's is, combat and noncombat alike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73e7a2ad-d11c-4867-b97d-f971809da778",
		Name:         "Glacial Chasm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Glacial Chasm — cumulative upkeep—pay 2 life", effects.LifePayment(2), "", ""),
			effects.WhenThisEnters("Glacial Chasm — sacrifice a land", effects.S82OldB614SacrificeALand("Glacial Chasm — sacrifice a land")),
		},
		Static: []game.StaticAbility{{
			Label: "Creatures you control can't attack.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack
			},
		}},
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToYou,
				Damage: effects.AnyDamage,
				Label:  "Glacial Chasm — prevent all damage that would be dealt to you",
			}),
		},
	})
}
