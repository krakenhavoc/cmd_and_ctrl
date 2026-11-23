package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Artificer's Epiphany — Instant {2}{U} (Jumpstart 2022, #5948):
//
//	"Draw two cards. If you control no artifacts, discard a card."
//
// Snow Day's draw-then-discard (the discard is the caster's choice, asked
// after the draws), made only when the caster controls no artifact as the
// spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed10bb4c-f7ef-4046-8dde-465041b55078",
		Name:         "Artificer's Epiphany",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve:    effects.GapB172ArtificersEpiphany,
	})
}
