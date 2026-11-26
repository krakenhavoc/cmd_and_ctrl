package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Howlpack Wolf — {2}{R} Creature — Wolf 3/3 (Shadows over Innistrad, S82 OLD
// b293):
//
//	"This creature can't block unless you control another Wolf or
//	 Werewolf."
//
// A CR 509.1b block rule on the creature itself; the Wolf or Werewolf is read
// through the layers as blockers are declared and must be another permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2fb7d03-d6bf-4716-9ca4-bd7dbcf9ae66",
		Name:         "Howlpack Wolf",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{effects.OldB293CantBlockUnlessYouControlAnother(
			"it can't block unless its controller controls another Wolf or Werewolf",
			effects.QuerySubtype("Wolf"), effects.QuerySubtype("Werewolf"))},
	})
}
