package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diabolic Tutor — Sorcery {2}{B}{B} (EDHREC rank 482):
//
//	"Search your library for a card, put that card into your hand,
//	 then shuffle."
//
// Demonic Tutor at twice the price — the same unconditional search
// with a nil predicate, so every card in the library is a candidate
// and the controller picks one through the S22 search chooser.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14589b6b-1814-46f9-a364-83cc15dacac2",
		Name:         "Diabolic Tutor",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneHand,
				Limit:   1,
				Shuffle: true,
				Reason:  "Diabolic Tutor — search your library for a card",
			}.Apply(ctx)
		},
	})
}
