package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Rats — Creature — Rat {1}{B}, 1/1 (Portal Second Age, #5625):
//
//	"When this creature enters, target opponent discards a card."
//
// Corrupt Court Official's entry: the opponent is targeted as the
// trigger goes on the stack and chooses their own discard (CR 701.9b).
// An opponent who has left by resolution is an illegal target and
// nothing happens (CR 608.2b); an empty hand discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fa1bbfd-92b5-482c-b32d-4cdc286474c4",
		Name:         "Ravenous Rats",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ravenous Rats — target opponent discards a card", effects.TargetOpponentDiscards(1)),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
