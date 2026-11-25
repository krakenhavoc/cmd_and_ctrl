package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Disappearing Act — Instant {1}{U}{U} (Kaladesh, #5539):
//
//	"As an additional cost to cast this spell, return a permanent you control
//	 to its owner's hand.
//	 Counter target spell."
//
// Familiar's Ruse's counterspell with its return cost over any permanent: the
// permanent goes back as the spell is cast (CR 601.2h), so it is paid even if
// the spell is then countered, and a player with no permanent can't cast it
// (CR 118.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "20393369-fdee-45ea-8584-451f38b1c002",
		Name:           "Disappearing Act",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: effects.ReturnPermanentToHandCost("a permanent you control"),
		Targets:        effects.TargetSpell("target spell"),
		OnResolve:      effects.CounterTheTargetSpell,
	})
}
