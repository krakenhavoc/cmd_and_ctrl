package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Personal Tutor — Sorcery {U} (Commander Masters, #5948):
//
//	"Search your library for a sorcery card, reveal it, then shuffle and put
//	 that card on top."
//
// Mystical Tutor's body (tutorToTop) over a sorcery card: shuffle first,
// then place the card on top, and the reveal shows the whole table what the
// next draw is. A sorcery card, not an instant: only the sorcery type counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90f54959-2c9b-4b8a-84c9-d6893eb43553",
		Name:         "Personal Tutor",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return tutorToTop(ctx, "Personal Tutor — a sorcery card", true,
				func(c game.Card) bool { return c.IsSorcery() })
		},
	})
}
