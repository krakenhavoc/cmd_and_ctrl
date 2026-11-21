package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantom Centaur — Creature — Centaur Spirit {2}{G}{G}, 2/0:
//
//	"Protection from black
//	 This creature enters with three +1/+1 counters on it.
//	 If damage would be dealt to this creature, prevent that damage. Remove a +1/+1 counter from this creature."
//
// ADR 0108 §8 (#1906): one of the Phantoms (phantoms.go). Protection from
// black is a prevention effect too, and the engine applies it first
// (ReplacementEffect.Preemptive), so a black source's damage costs no
// counter — the order the ruling lets its controller choose.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9bb8c54b-1228-4b7c-8651-52cb5b0f6e72",
		Name:            "Phantom Centaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
		Replacements:    effects.PhantomReplacements("Phantom Centaur", 3),
	})
}
