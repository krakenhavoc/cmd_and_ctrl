package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Withering Boon — Instant {1}{B} (Mirage, S82 OLD batch 812, #5633):
//
//	"As an additional cost to cast this spell, pay 3 life.
//	 Counter target creature spell."
//
// A mandatory life cost paid with the cast (CR 601.2h, refused below 3 life,
// CR 119.4) and Bone to Ash's counter over a creature spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "6d87415b-0618-48e1-9510-320c4611ecda",
		Name:           "Withering Boon",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: effects.PayLifeCost(3),
		Targets:        effects.TargetSpell("target creature spell", effects.Creature()),
		OnResolve:      effects.CounterTheTargetSpell,
	})
}
