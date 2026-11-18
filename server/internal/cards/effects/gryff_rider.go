package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gryff Rider — Creature — Human Knight {2}{W}, 2/1 (Crimson Vow):
//
//	"Flying
//	 Training (Whenever this creature attacks with another creature with
//	 greater power, put a +1/+1 counter on this creature.)"
//
// Flying rides PrintedKeywords; training is the shared Training() trigger
// (training.go, CR 702.149a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e1ca5c24-bbd8-4eed-bb2b-17163f90e72b",
		Name:            "Gryff Rider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{Training("Gryff Rider — training: put a +1/+1 counter on it")},
	})
}
