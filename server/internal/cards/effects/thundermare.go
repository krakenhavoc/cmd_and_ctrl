package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thundermare — Creature — Elemental Horse {5}{R}, 5/5 (Portal, #5630):
//
//	"Haste (This creature can attack and {T} as soon as it comes under your control.)
//	When this creature enters, tap all other creatures."
//
// Shrieking Mogg's body and enters trigger, word for word.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "58fa33fd-02a4-4ec9-9226-82594cfec363",
		Name:            "Thundermare",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Thundermare — tap all other creatures", tapAllOtherCreatures),
		},
	})
}
