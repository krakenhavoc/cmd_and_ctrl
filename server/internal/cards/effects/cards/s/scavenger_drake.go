package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scavenger Drake — {3}{B} Creature — Drake 1/1 (Shards of Alara, S82 OLD
// b514):
//
//	"Flying
//	 Whenever another creature dies, you may put a +1/+1 counter on this
//	 creature."
//
// Deathbringer Thoctar's trigger (any creature's death, including an
// opponent's, but not this one's), optional, on the shared body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7d51b860-f7a4-4b13-a83c-134931ae7baf",
		Name:            "Scavenger Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{effects.OldB514AnotherCreatureDiedMayCounter("Scavenger Drake")},
	})
}
