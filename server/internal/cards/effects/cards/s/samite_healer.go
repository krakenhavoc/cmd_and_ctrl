package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Samite Healer — Creature — Human Cleric {1}{W}, 1/1 (Eighth Edition, S82 OLD
// batch 612, #5607):
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target this turn."
//
// Benevolent Ancestor's ability (PreventNextOneDamageToTheTarget): a one-use
// shield of 1 on the target that lasts until end of turn (CR 615.7), charged as
// it prevents. A tap ability, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95a0ca48-d924-47f4-86ed-42c673ee778c",
		Name:         "Samite Healer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.TapCost(),
			Targets: effects.TargetAny(),
			Effect:  effects.PreventNextOneDamageToTheTarget("Samite Healer"),
		}},
	})
}
