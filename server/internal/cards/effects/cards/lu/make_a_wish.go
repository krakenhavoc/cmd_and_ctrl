package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Make a Wish — {3}{G} Sorcery (Innistrad, S82 OLD b447, #5572):
//
//	"Return two cards at random from your graveyard to your hand."
//
// The two cards are drawn without replacement from a persisted random stream
// (so an undo replays the same pick); a graveyard of fewer than two returns
// what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "888ccda6-896f-40f3-93a1-efc48254af34",
		Name:         "Make a Wish",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve:    effects.OldB447ReturnTwoAtRandom,
	})
}
