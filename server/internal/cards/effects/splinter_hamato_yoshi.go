package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splinter, Hamato Yoshi — Legendary Creature — Mutant Ninja Rat {1}{B},
// 1/3 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Sneak {B} (You may cast this spell for {B} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step. He enters tapped and attacking.)
//	 Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 Other Ninjas you control get +1/+1."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190); menace is
// engine-side; the anthem is the shared lord (TribalAnthem), which reads
// "Ninja" through Splinter's own text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "41fbdaa2-7a7a-44d5-a8f4-77624350eedd",
		Name:             "Splinter, Hamato Yoshi",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"menace"},
		AlternativeCosts: []game.AlternativeCost{Sneak("{B}")},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Ninja"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
