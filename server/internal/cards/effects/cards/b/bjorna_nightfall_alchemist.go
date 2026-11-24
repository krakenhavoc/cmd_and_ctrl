package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bjorna, Nightfall Alchemist — Legendary Creature — Human {U}{R}, 1/3 (Secret
// Lair Drop):
//
//	"{T}, Sacrifice an artifact: Bjorna, Nightfall Alchemist deals 1 damage to
//	 target creature. Goad that creature.
//	 Partner—Friends forever (You can have two commanders if both have this
//	 ability.)"
//
// The cost is a tap and the sacrifice of an artifact (paid as the ability is
// activated, CR 602.2b). The damage is the targeted one-damage ability's, and
// the goad is Disrupt Decorum's targeted goad (GoadTarget: a creature the damage
// killed, or one gone from the battlefield, is not goaded, CR 608.2b), lasting
// until the goader's next turn (CR 701.15a). Partner—Friends forever is a
// deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c880fbdc-bdd9-4f80-81d4-e3e1124f76ca",
		Name:         "Bjorna, Nightfall Alchemist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice an artifact: Bjorna, Nightfall Alchemist deals 1 damage to target creature. Goad that creature.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.DamageToFirstTarget(1)(g, item); err != nil {
					return err
				}
				return effects.GoadTarget(g, item)
			},
		}},
	})
}
