package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Airdrop Condor — Creature — Bird {4}{R}, 2/2 (Onslaught):
//
//	"Flying
//	 {1}{R}, Sacrifice a Goblin creature: This creature deals damage
//	 equal to the sacrificed creature's power to any target."
//
// Fling's damage as an activated ability: the Goblin is sacrificed as
// the cost (CR 602.2b), and its power is read as it last existed on the
// battlefield (CR 608.2h, Context.SacrificedPower), floored at zero. The
// Condor is the source of the damage. A sacrifice cost does not target,
// so hexproof on the Goblin never matters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "25fc274f-d5af-4174-b8e8-c2e65bb4c13f",
		Name:            "Airdrop Condor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice a Goblin creature: This creature deals damage equal to the sacrificed creature's power to any target.",
			Cost:    effects.Plus(effects.ManaCost("{1}{R}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Goblin creature", effects.OfCreatureType("Goblin"))}),
			Targets: effects.TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DamageEqualToSacrificedPower(item, effects.NewContext(g, item))
			},
		}},
	})
}
