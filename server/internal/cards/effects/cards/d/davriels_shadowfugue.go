package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Davriel's Shadowfugue — Sorcery {3}{B} (War of the Spark, #5520):
//
//	"Target player discards two cards and loses 2 life."
//
// Mind Rot's discard (the target chooses the two cards) and a life loss, which
// is not damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb490f3f-74ed-4d38-8eda-9948c095f3bc",
		Name:         "Davriel's Shadowfugue",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.S82OldB172TargetPlayerDiscardsTwoAndLosesTwo,
	})
}
