package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nullify — Instant {U}{U} (Born of the Gods,
// #5559):
//
//	"Counter target creature or Aura spell."
//
// Annul's body over a creature or Aura spell. A bestowed spell is an
// Aura spell and its card a creature card, so either clause takes it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ea295af-e117-49cc-836b-da16b5d5bbdf",
		Name:         "Nullify",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target creature or Aura spell", effects.Or(effects.Creature(), effects.Subtype("Aura"))),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
