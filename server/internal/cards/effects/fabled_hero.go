package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fabled Hero — Creature — Human Soldier {1}{W}{W}, 2/2 (Theros, #5561):
//
//	"Double strike
//	 Heroic — Whenever you cast a spell that targets this creature, put a +1/+1 counter on it."
//
// The shared heroic trigger (CR 207.2c). The counter lands only on the same object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f9fd9b1c-4c00-4db7-bbfe-7fd5d8bb9258",
		Name:            "Fabled Hero",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			heroic("Fabled Hero — a +1/+1 counter on it (heroic)", thisGetsAPlusOneCounter),
		},
	})
}
