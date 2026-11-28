package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tahngarth's Glare — Sorcery {R} (Apocalypse, #5614):
//
//	"Look at the top three cards of target opponent's library, then put them
//	 back in any order. That player looks at the top three cards of your
//	 library, then puts them back in any order."
//
// Two placements, each by its own looker over its own library (Second
// Sight's look at an opponent's library, and Fortune's Favor's opponent
// looking at yours): you order the opponent's top three, and that opponent
// orders yours. An empty library asks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "390cbc61-6291-4e7d-b262-4de57c827702",
		Name:         "Tahngarth's Glare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve:    effects.S82OldB635TahngarthsGlare,
	})
}
