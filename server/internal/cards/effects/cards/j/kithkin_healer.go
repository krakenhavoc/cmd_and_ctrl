package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kithkin Healer — Creature — Kithkin Cleric {2}{W}, 2/2 (Lorwyn, S82 OLD
// b547, #5593):
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target this
//	 turn."
//
// Benevolent Ancestor's prevention shield, without defender: a one-damage
// shield on the chosen creature, player or planeswalker, lasting until the
// turn ends (CR 615.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e3bd74f3-1f38-4162-a0be-c76ad63e2d9b",
		Name:         "Kithkin Healer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextOneDamageToTheTarget("Kithkin Healer"),
		}},
	})
}
