package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pale Wayfarer — Creature — Spirit Giant {5}{W}{W}, 4/4 (Shadowmoor, S82 OLD
// b531, #5591):
//
//	"{2}{W}{W}, {Q}: Target creature gains protection from the color of its
//	 controller's choice until end of turn. ({Q} is the untap symbol.)"
//
// Mother of Runes' protection grant, with two differences: the cost is {Q} (CR
// 107.6, AbilityCost.UntapSelf) and the colour is chosen by the target creature's
// controller, not the Wayfarer's, so a Wayfarer pointed at an opponent's creature
// asks that opponent. The colour is asked as the ability resolves, after the
// target is chosen, and a target that has left does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06b8acf2-3f2f-4e8f-9584-8d2ffa18a496",
		Name:         "Pale Wayfarer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{W}, {Q}: Target creature gains protection from the color of its controller's choice until end of turn.",
			Cost:    effects.OldB531UntapCost("{2}{W}{W}"),
			Purpose: game.Purpose{Answers: game.AnswerProtect, Targets: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection")).Targets},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB531TargetCreatureProtectionOfItsControllersChoice("Pale Wayfarer"),
		}},
	})
}
