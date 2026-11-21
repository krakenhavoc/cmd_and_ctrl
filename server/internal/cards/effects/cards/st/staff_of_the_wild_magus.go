package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Staff of the Wild Magus — Artifact {3} (Magic 2014, #5563):
//
//	"Whenever you cast a green spell or a Forest you control enters, you gain 1 life."
//
// One ability with two trigger conditions (m14StaffOfTheMagus): casting a
// green spell, and a Forest entering under its controller's control. Each
// is its own trigger, and each gains 1 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bdff38d-0a52-42e5-86ea-edd7029b44f6",
		Name:         "Staff of the Wild Magus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.M14StaffOfTheMagus("Staff of the Wild Magus", "G", "Forest"),
		},
	})
}
