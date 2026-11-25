package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whims of the Fates — Sorcery {5}{R} (Commander 2016, #5537):
//
//	"Starting with you, each player separates all permanents they control into
//	 three piles. Then each player chooses one of their piles at random and
//	 sacrifices those permanents. (Piles can be empty.)"
//
// SeparateIntoPiles (#6248, CR 700.3) with SacrificeARandomPileEach: every
// split is made before any pile is chosen, and the chosen piles are sacrificed
// together (CR 701.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d359434-eb1b-478e-a745-10e2c564d89f",
		Name:         "Whims of the Fates",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SeparateIntoPiles{Piles: 3, Then: effects.SacrificeARandomPileEach}.Apply(ctx)
		},
	})
}
