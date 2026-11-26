package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fierce Invocation — {4}{R} Sorcery (Fate Reforged):
//
//	"Manifest the top card of your library, then put two +1/+1 counters on it."
//
// S82OldB337ManifestThenCounters: the counters go on from the entry's tail.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e75fc29-68e3-41a9-9f59-adb4460663b2",
		Name:         "Fierce Invocation",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.S82OldB337ManifestThenCounters{N: 2}),
	})
}
