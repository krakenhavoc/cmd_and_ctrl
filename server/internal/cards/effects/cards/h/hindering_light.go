package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hindering Light — Instant {W}{U}:
//
//	"Counter target spell that targets you or a permanent you control.
//	 Draw a card."
//
// The clause reads the candidate spell's announced targets: the caster,
// or a permanent the caster controls on the battlefield. If the spell
// has gone by resolution the target is illegal and nothing happens, so
// no card is drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "945ac15c-780b-44a9-8a8e-011cf1c30876",
		Name:         "Hindering Light",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target spell that targets you or a permanent you control", effects.HinderingLightTargetsYouOrYourPermanent),
		OnResolve:    effects.CounterFirstTargetSpellThenDraw,
	})
}
