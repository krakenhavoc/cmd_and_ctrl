package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Familiar's Ruse — Instant {U}{U} (Lorwyn, #5593):
//
//	"As an additional cost to cast this spell, return a creature you
//	 control to its owner's hand.
//	 Counter target spell."
//
// Counterspell with Fear of Isolation's return cost: the creature goes back
// as the spell is cast (CR 601.2h), so it is paid even if the spell is then
// countered, and a player with no creature can't cast it (CR 118.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "3b6d773b-e2f7-4e74-9e77-f2e0d369a31e",
		Name:           "Familiar's Ruse",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: effects.ReturnPermanentToHandCost("a creature you control", effects.Creature()),
		Targets:        effects.TargetSpell("target spell"),
		OnResolve:      effects.CounterTheTargetSpell,
	})
}
