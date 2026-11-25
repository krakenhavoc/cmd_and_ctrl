package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invade the City — Sorcery {1}{U}{R} (War of the Spark, #5520):
//
//	"Amass Zombies X, where X is the number of instant and sorcery cards in
//	 your graveyard."
//
// X is counted as the spell resolves. Amass with X of 0 still creates the
// 0/0 Army when you have none (CR 701.47a), which then dies as a state-based
// action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "583a4744-89fb-4bd6-8d4c-c3d845e2e19c",
		Name:         "Invade the City",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := len(effects.GyCardIDs(ctx.Game, ctx.Controller(), effects.S82OldB172IsInstantOrSorcery))
			return effects.Amass{Subtype: "Zombie", N: x}.Apply(ctx)
		},
	})
}
