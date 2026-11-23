package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dawning Angel — Creature — Angel {4}{W}, 3/2 (Core Set 2020, #5518):
//
//	"Flying
//	 When this creature enters, you gain 4 life."
//
// Angel of Mercy's shape: the enters trigger uses the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7d88510-d2ce-4dd6-938c-964cc74ff7d7",
		Name:            "Dawning Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dawning Angel — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
