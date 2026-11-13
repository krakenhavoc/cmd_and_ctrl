package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spike Colony — Creature — Spike {4}{G}, 0/0 (Stronghold, #5627):
//
//	"This creature enters with four +1/+1 counters on it.
//	 {2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter
//	 on target creature."
//
// Spike Worker at four counters: the cycle's shared entry replacement
// (CR 614.1c) and counter move (spike_counters.go).
//
// No simplification.
func init() {
	const name = "Spike Colony"
	Register(Spec{
		OracleID:     "15902a15-090a-44e6-b8d2-59e0f26ed46f",
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{spikeEntersWith(name, "four", 4)},
		Activated:    []ActivatedAbility{spikeMovesACounter()},
	})
}
