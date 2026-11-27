package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Western Paladin — Creature — Phyrexian Zombie Knight {2}{B}{B}, 3/3
// (Eighth Edition, #5607):
//
//	"{B}{B}, {T}: Destroy target white creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2cbdbf9-9b08-419e-bb3e-482972d58e74",
		Name:         "Western Paladin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{B}, {T}: Destroy target white creature.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{B}{B}"), effects.TapCost()),
			Targets: effects.TargetCreature("target white creature", effects.OfColor("W")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
