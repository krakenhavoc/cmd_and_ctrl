package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sterling Supplier — Creature — Bird Soldier {4}{W}, 3/4 (Outlaws of
// Thunder Junction, #4468):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on another target
//	 creature you control."
//
// Ascendant Dustspeaker's enters trigger: "another" is object identity
// (Another), so the Supplier can't target itself, and with no other
// creature of yours the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Sterling Supplier — a +1/+1 counter on another target creature you control",
		PlusOneCounterOnTheTarget)
	enters.Targets = Another(TargetCreature("another target creature you control", YouControl()))
	Register(Spec{
		OracleID:        "ab23cca2-b59e-4d34-80de-98bd08985275",
		Name:            "Sterling Supplier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
