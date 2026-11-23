package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tribble Cluster — Creature — Tribble {2}{W}, 1/1 (Star Trek Commander,
// #4974):
//
//	"At the beginning of combat on your turn, put a +1/+1 counter on this
//	 creature.
//	 When this creature dies, create X 1/1 white Tribble creature tokens
//	 with "This token can't block," where X is this creature's power."
//
// The counter is CounterOnThis at the beginning of combat on its
// controller's turn. X is the Cluster's power as it last existed on the
// battlefield (CR 603.10a), so the counters it grew count. The Tribble is
// a catalog token with its can't-block static.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c1caf0f-9ee8-4eaa-b8e7-309ba9be9282",
		Name:         "Tribble Cluster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Tribble Cluster — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			effects.WhenThisDies("Tribble Cluster — create X 1/1 white Tribble tokens that can't block, where X is its power",
				effects.TrekB33TokensEqualToItsPower(effects.TribbleToken)),
		},
	})
}
