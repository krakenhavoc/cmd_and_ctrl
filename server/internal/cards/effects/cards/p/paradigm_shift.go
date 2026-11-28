package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Paradigm Shift — Sorcery {1}{U} (Weatherlight, S82 older sets batch 779, #5629):
//
//	"Exile all cards from your library. Then shuffle your graveyard into your library."
//
// The library is exiled as one event; only then (CR 608.2c) every card in the graveyard goes into the library and it is
// shuffled (CR 701.24), the graveyard as it is once the exile has finished.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7179c2c5-c59c-4793-a3b5-28d26d755157",
		Name:         "Paradigm Shift",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB779ParadigmShift(ctx)
		},
	})
}
