package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunderclap — Instant {2}{R} (Mercadian Masques, #5619):
//
//	"You may sacrifice a Mountain rather than pay this spell's mana
//	 cost.
//	 Thunderclap deals 3 damage to target creature."
//
// The shared damageToFirstTarget body at 3 on a creature, with Crash's
// one-Mountain sacrifice alternative cost (SacrificeInstead, #1727),
// paid with the spell already on the stack (CR 601.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02f5ceb4-8b9c-4107-972b-6eec365c4a6c",
		Name:         "Thunderclap",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		AlternativeCosts: []game.AlternativeCost{
			effects.SacrificeInstead(1, "a Mountain", 0, effects.HasSubtype("Mountain")),
		},
		OnResolve: effects.DamageToFirstTargetStep(3),
	})
}
