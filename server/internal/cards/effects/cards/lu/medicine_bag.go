package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Medicine Bag — Artifact {3} (Exodus, S82 OLD b737, #5626):
//
//	"{1}, {T}, Discard a card: Regenerate target creature."
//
// Rushwood Herbalist's regeneration shield (CR 701.19a) on any creature,
// paid with mana, tap and a discard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5801f773-e214-413d-968d-b84bf99f0642",
		Name:         "Medicine Bag",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Discard a card: Regenerate target creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.DiscardACard()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.RegenerateTheTargetPermanent,
		}},
	})
}
