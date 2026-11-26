package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Algae Gharial — {3}{G} Creature — Crocodile 1/1 (Shards of Alara, S82 OLD
// b514):
//
//	"Shroud (This creature can't be the target of spells or abilities.)
//	 Whenever another creature dies, you may put a +1/+1 counter on this
//	 creature."
//
// Scavenger Drake's trigger on a shrouded body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7020dc51-8271-40a3-9c4c-ebd40c0b08b2",
		Name:            "Algae Gharial",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Triggered:       []game.TriggeredAbility{effects.OldB514AnotherCreatureDiedMayCounter("Algae Gharial")},
	})
}
