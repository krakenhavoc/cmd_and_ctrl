package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Slayer — Creature — Phyrexian Minion {3}{B}, 2/2 (Invasion, S82 #5616):
//
//	"Flying
//	 Whenever this creature becomes blocked by a white creature, destroy that
//	 creature. It can't be regenerated."
//
// A per-blocker trigger (CR 509.3d): one for each white creature that blocks it,
// naming that blocker. The blocker is destroyed, ignoring regeneration (CR
// 701.19c), if that same object is still on the battlefield as the trigger
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "da74cb57-f6de-4d05-a339-6412beb8df25",
		Name:            "Phyrexian Slayer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, effects.BecomesBlockedByAWhiteCreature,
				"Phyrexian Slayer — destroy the white creature blocking it; it can't be regenerated", effects.DestroyTheBlockerNoRegen),
		},
	})
}
