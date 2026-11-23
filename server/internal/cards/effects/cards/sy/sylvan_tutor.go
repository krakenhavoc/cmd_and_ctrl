package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sylvan Tutor — Sorcery {G} (Special Guests, S82 BNS):
//
//	"Search your library for a creature card, reveal it, then shuffle and
//	 put that card on top."
//
// Worldly Tutor's text on a sorcery: the same shared tutorToTop body
// (one_mana_tutors.go), reveal included. The card stays in the library, so
// collecting it still costs a draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "935e0cac-51ee-4cb7-a209-f085e0f099ed",
		Name:         "Sylvan Tutor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TutorToTop(ctx, "Sylvan Tutor — a creature card", true,
				func(c game.Card) bool { return c.IsCreature() })
		},
	})
}
