package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hushwing Gryff — {2}{W} Creature — Hippogriff 2/1:
//
//	"Flash
//	 Flying
//	 Creatures entering don't cause abilities to trigger."
//
// Torpor Orb on a flash flier (#1735). Flash is how it is played: cast
// in response to a creature spell, it is on the battlefield when that
// creature enters, and the creature's enters triggers never happen.
func init() {
	effects.Register(effects.Spec{
		OracleID:           "a0d615b3-7ee3-43d7-ad37-5f7f1544a8db",
		Name:               "Hushwing Gryff",
		Completeness:       effects.CompletenessFull,
		PrintedKeywords:    []string{"flash", "flying"},
		TriggerSuppressors: []game.TriggerSuppressor{effects.CreaturesEnteringDontTrigger("Hushwing Gryff")},
	})
}
