package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Global Ruin — Sorcery {4}{W} (Invasion, S82 older sets batch 662, #5616):
//
//	"Each player chooses from the lands they control a land of each basic
//	 land type, then sacrifices the rest."
//
// Cataclysmic Gearhulk's ordered pick with a land rule: each player in turn
// order picks the lands to keep, covering every basic land type among their
// lands (a land with two types may be chosen for either, so a Dual alone can
// stand for both), and then every other land they control goes at once. A land
// with no basic land type is never kept. A board that needs no choice asks
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a7b89b2-5fe9-4d39-9d42-3ca78bcf6d40",
		Name:         "Global Ruin",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.OldB662GlobalRuin(ctx.Game, item)
		},
	})
}
