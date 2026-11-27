package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// See Beyond — Sorcery {1}{U} (Planechase 2012, S82 OLD, #5569):
//
//	"Draw two cards, then shuffle a card from your hand into your library."
//
// Brainstorm's draw-then-pick: the pick is raised after the draws, from the
// hand they left (CR 608.2c), and the chosen card is put into its owner's
// library and the library shuffled (a commander gets the CR 903.9 offer on the
// way). An empty hand asks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f9942ee-d3fd-4c01-93eb-870d642ebd83",
		Name:         "See Beyond",
		Purpose:      game.Purpose{Draws: 2, Answers: game.AnswerValue},
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.S82OldB431SeeBeyond,
	})
}
