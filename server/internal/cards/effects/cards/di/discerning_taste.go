package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Discerning Taste — Sorcery {2}{B} (Modern Horizons 2):
//
//	"Look at the top four cards of your library. Put one of them into your hand
//	 and the rest into your graveyard. You gain life equal to the greatest power
//	 among creature cards put into your graveyard this way."
//
// A look, so only you see the four. The pick is one card (mandatory while any
// are there); the rest are put into the graveyard, not milled. The life is the
// greatest power printed among the creature cards that really reached the
// graveyard, after the move; none, or only power 0, gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1896c513-ec48-4c8f-9acb-30bc6357fce2",
		Name:         "Discerning Taste",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve:    effects.OldB034DiscerningTaste,
	})
}
