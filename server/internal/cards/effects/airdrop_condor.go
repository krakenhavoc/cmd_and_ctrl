package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "25fc274f-d5af-4174-b8e8-c2e65bb4c13f",
		Name:            "Airdrop Condor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, Sacrifice a Goblin creature: This creature deals damage equal to the sacrificed creature's power to any target.",
			Cost:    Plus(ManaCost("{1}{R}"), game.AbilityCost{SacrificeOther: sacrificeSpec("a Goblin creature", OfCreatureType("Goblin"))}),
			Targets: TargetAny(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return damageEqualToSacrificedPower(item, NewContext(g, item))
			},
		}},
	})
}
