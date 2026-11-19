package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syndic of Tithes — Creature — Human Cleric {1}{W}, 2/2 (Ravnica: Clue Edition, #5929): "Extort."
//
// Extort is the shared keyword trigger (extort.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f5c52d2-0583-4c94-a043-15b2118877fa",
		Name:         "Syndic of Tithes",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Extort("Syndic of Tithes")},
	})
}
