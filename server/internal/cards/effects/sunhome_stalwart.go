package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunhome Stalwart — Creature — Human Soldier {1}{W}, 2/2 (Ravnica: Clue Edition, #5929): "First strike. Mentor."
//
// Mentor is the shared keyword trigger (source_relative_targets.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "340d6218-d351-4b2e-80da-7d4afa77190a",
		Name:            "Sunhome Stalwart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered:       []game.TriggeredAbility{Mentor("Sunhome Stalwart — mentor")},
	})
}
