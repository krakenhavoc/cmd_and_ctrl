package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cloudspire Skycycle — Artifact — Vehicle {2}{R}{W}, 2/3 (Aetherdrift,
// #4464):
//
//	"Flying
//	 When this Vehicle enters, distribute two +1/+1 counters among one or
//	 two other target Vehicles and/or creatures you control.
//	 Crew 1"
//
// Armament Dragon's distribution (Dividing, PutDividedCounters) as an
// enters trigger, one target fewer, over Vehicles and creatures you
// control other than the Skycycle itself (Another is object identity).
// An uncrewed Vehicle may take counters; they matter once it is a
// creature. The split is announced with the targets (CR 601.2d), and a
// target that has gone by resolution gets nothing (CR 608.2b). With no
// other creature or Vehicle you control, the trigger is removed for want
// of a target (CR 603.3d). Flying is the Vehicle's whether or not it is
// crewed; crew is the engine's (CR 702.122a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f7467fce-61e4-401a-8a3d-b0a2bc347088",
		Name:            "Cloudspire Skycycle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Cloudspire Skycycle — distribute two +1/+1 counters among one or two other target Vehicles and/or creatures you control",
				func(g *game.Game, item *game.StackItem) error {
					return PutDividedCounters(NewContext(g, item), game.CounterPlusOne)
				}),
				Another(TargetPermanent("one or two other target Vehicles and/or creatures you control",
					Or(Creature(), HasSubtype("Vehicle")), YouControl()).WithCount(1, 2).Dividing(Divide(2)))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Cloudspire Skycycle"),
		}},
	})
}
