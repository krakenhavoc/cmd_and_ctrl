package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Welkin Guide — {4}{W} Creature — Bird Cleric 2/2 (Eternal Masters, S82 OLD,
// #5542):
//
//	"Flying
//	 When this creature enters, target creature gets +2/+2 and gains flying
//	 until end of turn."
//
// Battleflight Eagle's enters trigger: the target, the Guide itself included,
// gets both halves if it is still legal as the trigger resolves (CR 608.2b) and
// neither if it is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0ce31ff1-6cfe-4b34-bd02-7d727934e754",
		Name:            "Welkin Guide",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Welkin Guide — target creature gets +2/+2 and gains flying until end of turn",
					effects.TrekB41PumpAndGrantEachTarget(2, 2, "Welkin Guide — +2/+2 and flying", "flying")),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "flying"))),
		},
	})
}
