package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Staff of the Flame Magus — Artifact {3} (Magic 2014, #5563):
//
//	"Whenever you cast a red spell or a Mountain you control enters, you gain 1 life."
//
// One ability with two trigger conditions (m14StaffOfTheMagus): casting a
// red spell, and a Mountain entering under its controller's control. Each
// is its own trigger, and each gains 1 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fc71710a-8294-4370-89da-7ab1c2b38d19",
		Name:         "Staff of the Flame Magus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.M14StaffOfTheMagus("Staff of the Flame Magus", "R", "Mountain"),
		},
	})
}
