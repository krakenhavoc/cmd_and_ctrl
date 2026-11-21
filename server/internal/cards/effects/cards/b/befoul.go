package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Befoul — Sorcery {2}{B}{B} (Urza's Saga):
//
//	"Destroy target land or nonblack creature. It can't be regenerated."
//
// One target clause, a land or a nonblack creature, judged at announce
// and again at resolution (CR 608.2b): a creature that turned black in
// response is no longer a legal target, and the spell does nothing. A
// black land creature is still a land, so it can be chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c573c59a-5a79-4fe1-a6d8-5e0153b5c059",
		Name:         "Befoul",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target land or nonblack creature", effects.Or(effects.Land(), effects.And(effects.Creature(), effects.NonBlack()))),
		OnResolve:    effects.DestroyTheTargetPermanentNoRegen,
	})
}
