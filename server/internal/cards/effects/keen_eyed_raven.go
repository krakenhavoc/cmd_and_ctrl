package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Keen-Eyed Raven — Creature — Bird {2}{W}, 2/2 (Assassin's Creed,
// #5466):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on another target
//	 creature you control."
//
// "Another" is object identity (Another excludes the source), so the
// Raven never offers itself; with no other creature the trigger is
// removed for want of a target (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e42450c7-a9de-412f-b439-76f2f9b6031b",
		Name:            "Keen-Eyed Raven",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Keen-Eyed Raven — put a +1/+1 counter on another target creature you control",
					PlusOneCounterOnChosenTarget),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
