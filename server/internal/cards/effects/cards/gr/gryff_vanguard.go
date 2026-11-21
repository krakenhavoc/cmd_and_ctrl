package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gryff Vanguard — {4}{U} Creature — Human Knight 3/2 (Avacyn Restored, #5570):
//
//	"Flying
//	 When this creature enters, draw a card."
//
// Flying and Skyscanner's enters draw (samiDrawOnETB), a trigger
// that uses the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3bcdf832-4ce2-4043-b46a-0c5fd7019fcd",
		Name:            "Gryff Vanguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.SamiDrawOnETB("Gryff Vanguard")},
	})
}
