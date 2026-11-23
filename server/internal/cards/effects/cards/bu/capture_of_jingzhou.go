package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Capture of Jingzhou — Sorcery {3}{U}{U}:
//
//	"Take an extra turn after this one."
//
// Temporal Manipulation's text under another name; see that file.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "89aa65d9-2502-40b0-90b6-b25a8e9f6155",
		Name:         "Capture of Jingzhou",
		Completeness: effects.CompletenessFull,
		// ADR 0188 Decision 10: the caster's extra turn, priced by the bot.
		Purpose:   game.Purpose{ExtraTurns: 1},
		OnResolve: effects.YouTakeAnExtraTurn,
	})
}
