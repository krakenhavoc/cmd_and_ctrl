package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flash — Instant {1}{U} (Mirage, S82 OLD b798, #5633):
//
//	"You may put a creature card from your hand onto the battlefield. If you
//	 do, sacrifice it unless you pay its mana cost reduced by {2}."
//
// Elvish Piper's put (a put, never a cast, with a real "you may"), then a
// registered rider told the permanent that arrived: it asks for its mana cost
// less {2} of generic mana (a cost of {0} is owed nothing) and sacrifices the
// creature on a decline or a payment that cannot be made. If the creature has
// left by then, nothing is asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e057471-e415-4105-9064-8236a0a5047d",
		Name:         "Flash",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerMakesBlocker},
		OnResolve:    effects.OldB798Flash,
	})
}
