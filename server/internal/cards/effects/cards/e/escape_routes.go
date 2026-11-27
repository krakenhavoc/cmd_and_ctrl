package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Escape Routes — Enchantment {2}{U} (Planeshift, S82 OLD b644, #5615):
//
//	"{2}{U}: Return target white or black creature you control to its owner's
//	 hand."
//
// The target is a white or black creature the activator controls, chosen as the
// ability is activated; it returns as the ability resolves if it is still legal
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65260087-48b5-4968-8584-7cc561bb312c",
		Name:         "Escape Routes",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Return target white or black creature you control to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCreature("target white or black creature you control", effects.And(effects.YouControl(), effects.Or(effects.OfColor("W"), effects.OfColor("B")))),
			Effect:  effects.BounceChosenTarget,
		}},
	})
}
