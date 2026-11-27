package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Noble Vestige — Creature — Spirit {2}{W}, 1/2 (Zendikar, #5583):
//
//	"Flying
//	 {T}: Prevent the next 1 damage that would be dealt to target player or
//	 planeswalker this turn."
//
// Jhessian Balmgiver's charged shield (CR 615.7) on one player or
// planeswalker; the {T} cost keeps summoning sickness on it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "63ec8167-2dc3-49a6-a7ed-6b1474bbda2c",
		Name:            "Noble Vestige",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to target player or planeswalker this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetPlayerOrPlaneswalker(),
			Effect:  effects.PreventNextOneDamageToTheTarget("Noble Vestige"),
		}},
	})
}
