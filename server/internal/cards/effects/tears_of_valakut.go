package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tears of Valakut — Instant {1}{R} (Oath of the Gatewatch, #5545):
//
//	"This spell can't be countered.
//	 Tears of Valakut deals 5 damage to target creature with flying."
//
// Abrupt Decay's CantBeCountered (CR 101.2) and a flying-only target
// clause; the damage is the spell's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fce335fb-bc67-4fa4-8d38-097f2a661332",
		Name:            "Tears of Valakut",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Purpose:         game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		Targets:         TargetCreature("target creature with flying", HasKeyword("flying")),
		OnResolve:       damageToFirstTarget(5),
	})
}
