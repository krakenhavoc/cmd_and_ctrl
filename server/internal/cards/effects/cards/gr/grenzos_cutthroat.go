package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grenzo's Cutthroat — Creature — Goblin Rogue {1}{R}, 1/1 (Conspiracy, S82
// OLD batch 363, #5557):
//
//	"First strike
//	 Dethrone (Whenever this creature attacks the player with the most life
//	 or tied for most life, put a +1/+1 counter on it.)"
//
// First strike is the printed keyword; dethrone is the shared GapB201Dethrone
// trigger (CR 702.105a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a133991-8a7a-4bfd-b0ae-d76433450c03",
		Name:            "Grenzo's Cutthroat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.GapB201Dethrone("Grenzo's Cutthroat — dethrone: put a +1/+1 counter on it"),
		},
	})
}
