package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunblast Angel — Creature — Angel {4}{W}{W}, 4/5:
//
//	"Flying
//	 When this creature enters, destroy all tapped creatures."
//
// Every creature tapped as the ability resolves, yours and the
// Angel itself included.
func init() {
	Register(Spec{
		OracleID:        "8b34d1df-4024-497d-bd28-a7caa8c3fa24",
		Name:            "Sunblast Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				WhenThisEnters("Sunblast Angel — destroy all tapped creatures",
					Do(DestroyAllMatching{Match: And(Creature(), tappedPermanent())})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
		},
	})
}
