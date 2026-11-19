package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreadmobile — Artifact — Vehicle {2}{B}, 3/3 (Modern Horizons 3, #5948):
//
//	"Menace
//	 {1}, Sacrifice another artifact or creature: Put a +1/+1 counter on this
//	 Vehicle.
//	 Crew 1"
//
// Menace is the printed keyword and crew is the shared Crew ability (the
// Vehicle is a creature only after it resolves, CR 301.7). The sacrifice is a
// cost naming another artifact or creature you control (CR 602.2b), so it is
// paid before the counter resolves: the counter is put on the Vehicle itself
// whether or not it is a creature then, and a Vehicle that has left does
// nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "51cc270d-55c8-4b6b-a6d7-f98d7b8e5185",
		Name:            "Dreadmobile",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, Sacrifice another artifact or creature: Put a +1/+1 counter on this Vehicle.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
				Cost:    Plus(ManaCost("{1}"), SacrificeAnotherN(1, "another artifact or creature", Or(Artifact(), Creature()))),
				Effect:  Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
			},
			{
				Label:   "Crew 1",
				Cost:    CrewCost(1),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Dreadmobile"),
			},
		},
	})
}
