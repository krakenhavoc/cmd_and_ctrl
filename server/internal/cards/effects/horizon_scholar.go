package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horizon Scholar — Creature — Sphinx {5}{U}, 4/4 (Theros, #5561):
//
//	"Flying
//	 When this creature enters, scry 2. (Look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)"
//
// Faerie Seer's shape: an enters trigger that scries 2 on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "29933a32-3738-4ec5-aa6d-ce3684e2e5a3",
		Name:            "Horizon Scholar",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Scry: 2},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Horizon Scholar — scry 2", Do(Scry{N: 2})),
		},
	})
}
