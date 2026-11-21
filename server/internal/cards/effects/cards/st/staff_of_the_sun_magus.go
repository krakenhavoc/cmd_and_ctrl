package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Staff of the Sun Magus — Artifact {3} (Magic 2014, #5563):
//
//	"Whenever you cast a white spell or a Plains you control enters, you gain 1 life."
//
// One ability with two trigger conditions (m14StaffOfTheMagus): casting a
// white spell, and a Plains entering under its controller's control. Each
// is its own trigger, and each gains 1 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd8bc927-e50a-4fab-9803-0602a841c7c2",
		Name:         "Staff of the Sun Magus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.M14StaffOfTheMagus("Staff of the Sun Magus", "W", "Plains"),
		},
	})
}
