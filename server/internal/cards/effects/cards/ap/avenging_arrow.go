package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avenging Arrow — Instant {2}{W} (Return to Ravnica, #5566):
//
//	"Destroy target creature that dealt damage this turn."
//
// The creature must have been the source of damage (combat or not) this turn;
// MshA6DealtDamageThisTurn reads it off the turn's damage events, as Treacherous
// Greed's cost does. Checked at announce and again on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c29add6-8524-4a33-9f2e-f792032dd6cd",
		Name:         "Avenging Arrow",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature that dealt damage this turn", effects.MshA6DealtDamageThisTurn()),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
