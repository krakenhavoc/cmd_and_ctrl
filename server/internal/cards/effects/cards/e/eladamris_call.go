package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eladamri's Call — Instant {G}{W} (EDHREC rank 770):
//
//	"Search your library for a creature card, reveal that card, put
//	 it into your hand, then shuffle."
//
// Instant-speed creature tutor. Type-filtered search to hand,
// revealed, through the S22 chooser.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4acb6612-54e8-428d-acb6-c7259a5ad6a8",
		Name:         "Eladamri's Call",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.B06TutorToHand("Eladamri's Call — a creature card", func(c game.Card) bool { return c.IsCreature() }),
	})
}
