package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Memoricide — Sorcery {3}{B} (Scars of Mirrodin, S82 OLD batch 471, #5577):
//
//	"Choose a nonland card name. Search target player's graveyard, hand, and
//	 library for any number of cards with that name and exile them. Then that
//	 player shuffles."
//
// Slaughter Games' extraction (OldB426SlaughterGames) aimed at any player
// rather than an opponent: the name is chosen as the spell resolves and is
// free text, so "nonland" is kept where it is read, and a target gone by
// resolution makes the spell do nothing and no name is asked (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "deeb942b-cf50-43f7-90a9-bdce4e42123a",
		Name:         "Memoricide",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.OldB426SlaughterGames("Memoricide"),
	})
}
