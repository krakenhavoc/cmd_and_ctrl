package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firefly — Creature — Insect {3}{R}, 1/1 (Tempest, #5628):
//
//	"Flying
//	 {R}: This creature gets +1/+0 until end of turn."
//
// Shivan Dragon's text on a 1/1: flying is a printed keyword, and the
// firebreathing is the shared thisGetsUntilEndOfTurn body at {R}, ending
// in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6b151418-092f-4a43-8b21-3de64ea80cab",
		Name:            "Firefly",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Firefly — +1/+0"),
		}},
	})
}
