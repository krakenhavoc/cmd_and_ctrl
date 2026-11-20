package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artillerize — Instant {3}{R} (New Phyrexia, #5575):
//
//	"As an additional cost to cast this spell, sacrifice an artifact or creature.
//	 Artillerize deals 5 damage to any target."
//
// The sacrifice is an additional cost paid at cast (CR 601.2h,
// SacrificeCost), and Lightning Bolt's damage at 5.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "d14a8d90-13c0-4eb0-ab7b-86edb2c7fb96",
		Name:           "Artillerize",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("an artifact or creature", Or(Artifact(), Creature())),
		Targets:        TargetAny(),
		Purpose:        game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve:      damageToFirstTarget(5),
	})
}
