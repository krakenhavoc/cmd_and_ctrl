package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loyal Tutor — Instant {W} (Reality Fracture):
//
//	"Search your library for a planeswalker card, reveal it, then
//	 shuffle and put that card on top."
//
// The white tutor of the one-mana cycle: the revealed planeswalker sits
// on an unknown library and costs a draw step to collect. The shared
// tutorToTop body places it after the shuffle.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e52f151-d1f6-4fcb-9b21-1baecae27da6",
		Name:         "Loyal Tutor",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TutorToTop(ctx, "Loyal Tutor — a planeswalker card", true,
				func(c game.Card) bool { return c.IsPlaneswalker() })
		},
	})
}
