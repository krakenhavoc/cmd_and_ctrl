package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syndicate Enforcer — Creature — Human Rogue {3}{B}, 3/2 (Ravnica: Clue Edition, #5929): "Extort."
//
// Extort is the shared keyword trigger (extort.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a067b98-e775-4b44-addf-9d31716c2191",
		Name:         "Syndicate Enforcer",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Extort("Syndicate Enforcer")},
	})
}
