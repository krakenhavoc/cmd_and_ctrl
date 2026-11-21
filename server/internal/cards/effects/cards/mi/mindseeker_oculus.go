package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindseeker Oculus — Creature — Homunculus {2}{U}, 2/1 (Reality
// Fracture):
//
//	"When this creature enters, empower Jace 4."
//
// ADR 0139 proof card: the keyword action from an enters trigger, on
// the stack, so an opponent can respond before the Jace appears.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aa9547b3-e4e4-4557-9dd1-3aa9a7cb9937",
		Name:         "Mindseeker Oculus",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mindseeker Oculus — empower Jace 4", effects.Do(effects.EmpowerJace{N: 4})),
		},
	})
}
