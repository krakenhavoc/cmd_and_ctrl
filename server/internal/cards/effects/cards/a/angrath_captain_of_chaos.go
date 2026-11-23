package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angrath, Captain of Chaos — Legendary Planeswalker — Angrath {2}{B/R}{B/R}, loyalty 5 (War of the Spark, #5520):
//
//	"Creatures you control have menace.
//	 −2: Amass Zombies 2. (Put two +1/+1 counters on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)"
//
// Starting loyalty is printed data (ADR 0032). The static grants menace to your creatures;
// the −2 is the Amass primitive.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de08f5d4-8868-4d0f-8c0e-d0d3a2105581",
		Name:         "Angrath, Captain of Chaos",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.B16CreaturesYouControl, "menace"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "−2: Amass Zombies 2.",
			Purpose: game.Purpose{Amass: 2},
			Cost:    effects.LoyaltyCost(-2),
			Effect:  effects.Do(effects.Amass{Subtype: "Zombie", N: 2}),
		}},
	})
}
