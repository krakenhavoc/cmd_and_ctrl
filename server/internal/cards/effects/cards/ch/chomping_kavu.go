package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chomping Kavu — Creature — Kavu {3}{G}, 3/3 (March of the Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target
//	 creature. If that's another creature, it gains the following ability
//	 until end of turn.)
//	 This creature can't be blocked by creatures with power 2 or less."
//
// The static is Bristlebane Outrider's block rule on the Kavu itself. What
// backup grants is the same restriction as a scoped block-rule mod
// (game.CantBeBlockedByPowerMod, the one Rhonas's Stalwart registers),
// judged as blockers are declared on each would-be blocker's current power
// (CR 509.1b). BackupAlso applies it with the counter, to another creature
// only (CR 702.165a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20a362c3-d278-437f-881c-e7d9fe78fc2d",
		Name:         "Chomping Kavu",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less"),
		},
		Triggered: []game.TriggeredAbility{
			effects.BackupAlso("Chomping Kavu", 1, "", []game.Mod{effects.CantBeBlockedByPowerMod(2)}),
		},
	})
}
