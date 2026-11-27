package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cataclysm — Sorcery {2}{W}{W} (Exodus, S82 OLD batch 742, #5626):
//
//	"Each player chooses from among the permanents they control an artifact, a
//	 creature, an enchantment, and a land, then sacrifices the rest."
//
// Cataclysmic Gearhulk's ordered pick with lands in and planeswalkers out: each
// player in turn order picks the permanents to keep, one for each of the four
// types they have, then everything else they control is sacrificed at once
// (CR 701.21a). A planeswalker or battle can't be kept. A board that needs no
// choice asks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0479ca4-e925-4aa7-a4fa-f7e01f899ac5",
		Name:         "Cataclysm",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB742Cataclysm(ctx.Game, item)
		},
	})
}
