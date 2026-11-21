package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunhome Stalwart — Creature — Human Soldier {1}{W}, 2/2 (Ravnica: Clue Edition, #5929): "First strike. Mentor."
//
// Mentor is the shared keyword trigger (source_relative_targets.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "340d6218-d351-4b2e-80da-7d4afa77190a",
		Name:            "Sunhome Stalwart",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered:       []game.TriggeredAbility{effects.Mentor("Sunhome Stalwart — mentor")},
	})
}
