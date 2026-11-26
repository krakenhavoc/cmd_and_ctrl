package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Formless Nurturing — {3}{G} Sorcery (Fate Reforged):
//
//	"Manifest the top card of your library, then put a +1/+1 counter on it."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ddbe6d9b-ea69-449c-a765-b3c16ff60c08",
		Name:         "Formless Nurturing",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.S82OldB337ManifestThenCounters{N: 1}),
	})
}
