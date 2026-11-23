package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overprotect — Instant {1}{G} (Bloomburrow, #4467):
//
//	"Target creature you control gets +3/+3 and gains trample, hexproof,
//	 and indestructible until end of turn."
//
// One target, a pump and three keywords until end of turn on the
// still-legal target (pumpAndGrantFirstTarget, Adamant Will's body); a
// target that left in response, or one you no longer control, counters
// the spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8c3a293-ae6d-4c52-be2d-2d4b6a8edb66",
		Name:         "Overprotect",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "trample", "hexproof", "indestructible")), game.AnswerProtect|game.AnswerSurvive|game.AnswerPump|game.AnswerCombatGrant),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(3, 3, "Overprotect — +3/+3, trample, hexproof, and indestructible",
				"trample", "hexproof", "indestructible")(ctx.Game, item)
		},
	})
}
