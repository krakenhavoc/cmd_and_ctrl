package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aven Redeemer — Creature — Bird Cleric {3}{W}, 2/2 (Legions, #5609):
//
//	"Flying
//	 {T}: Prevent the next 2 damage that would be dealt to any target this
//	 turn."
//
// Benevolent Ancestor's shield at 2: a one-use prevention of 2 on the
// target that lasts until end of turn and is used up as it prevents
// (CR 615.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "faac29f5-10ad-4800-a9d3-2259c4800352",
		Name:            "Aven Redeemer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 2 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextDamageToTheTarget("Aven Redeemer", 2),
		}},
	})
}
