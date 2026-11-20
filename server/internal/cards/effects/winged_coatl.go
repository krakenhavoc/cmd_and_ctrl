package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winged Coatl — Creature — Snake {1}{G}{U}, 1/1 (Alara Reborn, #5586):
//
//	"Flash
//	 Flying
//	 Deathtouch"
//
// Three printed keywords the engine enforces. Flash makes it a blocker
// cast in the opponent's combat (Cathar Commando's declaration).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "11369d8b-6443-43c3-a60f-dd920a4e8c08",
		Name:            "Winged Coatl",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Answers: game.AnswerMakesBlocker},
		PrintedKeywords: []string{"flash", "flying", "deathtouch"},
	})
}
