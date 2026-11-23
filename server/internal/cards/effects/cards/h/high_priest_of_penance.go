package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// High Priest of Penance — {W}{B} Creature — Human Cleric 1/1
//
//	"Whenever this creature is dealt damage, you may destroy target nonland
//	 permanent."
//
// One trigger per batch of damage (CR 510.2, the enrage shape). The target is
// chosen as the trigger goes on the stack and re-checked as it resolves
// (CR 608.2b); the "may" is the controller's answer as it goes on the stack.
func init() {
	trigger := effects.Cmdr2B08EnrageTrigger("High Priest of Penance — you may destroy target nonland permanent", effects.DestroyFirstLegalTarget)
	effects.Register(effects.Spec{
		OracleID:     "d8070ee7-dbd1-4d38-9715-ba14e103e379",
		Name:         "High Priest of Penance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(trigger, effects.TargetPermanent("target nonland permanent", effects.Nonland())),
				"High Priest of Penance — destroy target nonland permanent?"),
		},
	})
}
