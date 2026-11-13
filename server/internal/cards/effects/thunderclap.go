package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "02f5ceb4-8b9c-4107-972b-6eec365c4a6c",
		Name:         "Thunderclap",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		AlternativeCosts: []game.AlternativeCost{
			SacrificeInstead(1, "a Mountain", 0, HasSubtype("Mountain")),
		},
		OnResolve: damageToFirstTarget(3),
	})
}
