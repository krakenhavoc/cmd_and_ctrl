package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hada Freeblade — Creature — Human Soldier Ally {W}, 0/1 (Worldwake, #5582):
//
//	"Whenever this creature or another Ally you control enters, you may
//	 put a +1/+1 counter on this creature."
//
// The Ally trigger is South Pole Voyager's: this creature's own entry,
// or another Ally entering under its controller's control
// (wwkThisOrAnotherAllyEntered). The +1/+1 counter is a CR 603.5 "you
// may" asked as the trigger resolves (wwkAllyMayCounter).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9ad9df4-2fd6-4a22-b861-0a8d28f2fe10",
		Name:         "Hada Freeblade",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.WwkAllyMayCounter("Hada Freeblade")},
	})
}
