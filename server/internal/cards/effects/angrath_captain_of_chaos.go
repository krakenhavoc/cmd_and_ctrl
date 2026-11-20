package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "de08f5d4-8868-4d0f-8c0e-d0d3a2105581",
		Name:         "Angrath, Captain of Chaos",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "menace"),
		},
		Activated: []ActivatedAbility{{
			Label:   "−2: Amass Zombies 2.",
			Purpose: game.Purpose{Amass: 2},
			Cost:    LoyaltyCost(-2),
			Effect:  Do(Amass{Subtype: "Zombie", N: 2}),
		}},
	})
}
