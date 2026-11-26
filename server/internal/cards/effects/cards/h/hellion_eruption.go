package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hellion Eruption — Sorcery {5}{R} (Duel Decks: Zendikar vs. Eldrazi,
// #5548):
//
//	"Sacrifice all creatures you control, then create that many 4/4 red
//	 Hellion creature tokens."
//
// The count is what was really sacrificed (a creature that can't be
// sacrificed makes no token), told by a keyed tail once the move has landed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1894a57-0faa-4e0d-8ab9-e4785f0457a4",
		Name:         "Hellion Eruption",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB316SacrificeThenTokens(ctx.Game, item)
		},
	})
}
