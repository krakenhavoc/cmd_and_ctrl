package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spike Worker — Creature — Spike {2}{G}, 0/0 (Stronghold, #5627):
//
//	"This creature enters with two +1/+1 counters on it.
//	 {2}, Remove a +1/+1 counter from this creature: Put a +1/+1 counter
//	 on target creature."
//
// The Spike cycle's shared entry replacement (CR 614.1c) at two and its
// shared counter move (spike_counters.go). The counters are what make it
// a 2/2 (CR 122.1a); moving the last one away leaves a 0/0 that dies.
//
// No simplification.
func init() {
	const name = "Spike Worker"
	Register(Spec{
		OracleID:     "d7fcf43a-191f-4f44-aade-0cf89079a0c5",
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{spikeEntersWith(name, "two", 2)},
		Activated:    []ActivatedAbility{spikeMovesACounter()},
	})
}
