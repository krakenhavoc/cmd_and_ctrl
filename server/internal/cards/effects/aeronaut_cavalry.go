package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aeronaut Cavalry — Creature — Human Soldier {4}{W}, 3/4 (The Brothers'
// War, #5482):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on another target
//	 Soldier you control."
//
// The target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b); "another" excludes the Cavalry itself.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Aeronaut Cavalry — a +1/+1 counter on another target Soldier you control",
		PlusOneCounterOnTheTarget)
	enters.Targets = Another(TargetCreature("another target Soldier you control", YouControl(), OfCreatureType("Soldier")))
	Register(Spec{
		OracleID:        "153292f9-e0d1-43a7-a083-275c49d0554a",
		Name:            "Aeronaut Cavalry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
