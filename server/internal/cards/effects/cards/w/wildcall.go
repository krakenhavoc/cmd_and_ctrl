package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wildcall — {X}{G}{G} Sorcery (Fate Reforged):
//
//	"Manifest the top card of your library, then put X +1/+1 counters on it."
//
// X is the value announced as the spell was cast (CR 107.3); X = 0 manifests
// with no counters.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d96ac790-428b-4a64-8dbd-6baa73eb6210",
		Name:         "Wildcall",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB337ManifestThenCounters{N: ctx.X()}.Apply(ctx)
		},
	})
}
