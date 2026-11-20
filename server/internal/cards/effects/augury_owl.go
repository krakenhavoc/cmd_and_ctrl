package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Augury Owl — Creature — Bird {1}{U}, 1/1 (Magic 2011, #5579):
//
//	"Flying
//	 When this creature enters, scry 3."
//
// Flying is the printed keyword. The enters trigger is Cavern Stomper's
// scry at three: the controller sorts the top three between top and
// bottom as it resolves (CR 701.22a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44c5f8e1-d9b8-4067-a60a-1ecc8bd11145",
		Name:            "Augury Owl",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Scry: 3},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Augury Owl — scry 3", Do(Scry{N: 3})),
		},
	})
}
