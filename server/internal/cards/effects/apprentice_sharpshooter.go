package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Apprentice Sharpshooter — Creature — Human Archer {2}{G}, 1/4 (Crimson Vow):
//
//	"Reach
//	 Training (Whenever this creature attacks with another creature with
//	 greater power, put a +1/+1 counter on this creature.)"
//
// Reach rides PrintedKeywords; training is the shared Training() trigger
// (training.go, CR 702.149a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40d6e1ff-995b-42fc-8146-8d5e5f1adcd9",
		Name:            "Apprentice Sharpshooter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{Training("Apprentice Sharpshooter — training: put a +1/+1 counter on it")},
	})
}
