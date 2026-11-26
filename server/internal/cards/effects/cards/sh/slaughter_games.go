package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slaughter Games — Sorcery {2}{B}{R} (Return to Ravnica, #5566):
//
//	"This spell can't be countered.
//	 Choose a nonland card name. Search target opponent's graveyard, hand, and
//	 library for any number of cards with that name and exile them. Then that
//	 player shuffles."
//
// Infinite Obliteration's extraction with "nonland" for "creature" (see
// OldB426SlaughterGames), and the spec's CantBeCountered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c38c7899-b774-489b-9eb7-084522635a3f",
		Name:            "Slaughter Games",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets:         effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		OnResolve:       effects.OldB426SlaughterGames("Slaughter Games"),
	})
}
