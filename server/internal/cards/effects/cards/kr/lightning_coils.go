package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Coils — Artifact {3} (Mirrodin, #5606):
//
//	"Whenever a nontoken creature you control dies, put a charge counter on
//	 this artifact.
//	 At the beginning of your upkeep, if this artifact has five or more charge
//	 counters on it, remove all of them from it and create that many 3/1 red
//	 Elemental creature tokens with haste. Exile them at the beginning of the
//	 next end step."
//
// Banshee's Blade's charge counter, on a creature of yours dying (a token is
// no card and does not count), and Valduk's Elementals with their exile at the
// next end step (CR 603.7). The upkeep trigger is an intervening "if"
// (CR 603.4): five counters when the upkeep begins and still as it resolves;
// every counter comes off and that many tokens are made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbb16f23-fdb1-444b-bef8-ce3e177c2678",
		Name:         "Lightning Coils",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.OldB607NontokenCreatureYouControlDied,
				"Lightning Coils — put a charge counter on this artifact", effects.B33PutChargeCounterOnSelf),
			effects.OldB607CoilsUpkeep("Lightning Coils — remove all charge counters and create that many 3/1 red Elemental creature tokens with haste; exile them at the beginning of the next end step"),
		},
	})
}
