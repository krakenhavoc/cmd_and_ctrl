package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spike Cannibal — Creature — Spike {1}{B}{B}, 0/0 (Exodus, S82 OLD batch 741, #5626):
//
//	"This creature enters with a +1/+1 counter on it.
//	 When this creature enters, move all +1/+1 counters from all creatures
//	 onto it."
//
// The Spike cycle's entry replacement (spike_counters.go), then a trigger
// that takes every +1/+1 counter off every creature (the Cannibal's own stay)
// and puts the total on it, if it is still the same permanent (CR 400.7);
// if it is gone nothing moves.
//
// No simplification.
func init() {
	const name = "Spike Cannibal"
	effects.Register(effects.Spec{
		OracleID:     "c175ab0e-7085-41ec-83e4-38fc85c338fa",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SpikeEntersWith(name, "a", 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spike Cannibal — move all +1/+1 counters from all creatures onto it", effects.S82OldB741MoveAllPlusOneOntoThis),
		},
	})
}
