package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raphael, the Nightwatcher — Legendary Creature — Mutant Ninja Turtle
// {2}{R}{R}, 2/3 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Sneak {1}{R}{R} (You may cast this spell for {1}{R}{R} if you also
//	 return an unblocked attacker you control to hand during the declare
//	 blockers step. He enters tapped and attacking.)
//	 Attacking creatures you control have double strike."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). The static
// is Blade Historian's (AttackingCreaturesHave), which also covers a
// creature put onto the battlefield attacking, Raphael himself included
// when he is sneaked in.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "8c1e5b25-ee5a-4ea9-a715-08dd1a4f51c4",
		Name:             "Raphael, the Nightwatcher",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{1}{R}{R}")},
		Static: []game.StaticAbility{
			AttackingCreaturesHave("Attacking creatures you control have double strike.", nil, "double strike"),
		},
	})
}
