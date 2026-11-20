package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inescapable Blaze — Instant {4}{R}{R} (Guilds of Ravnica, #5525):
//
//	"This spell can't be countered.
//	 Inescapable Blaze deals 6 damage to any target."
//
// Lightning Bolt's body for 6, with Abrupt Decay's CantBeCountered: a
// counter effect does nothing to it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5fcc057e-e6c2-4e37-a786-9a5ef94926e1",
		Name:            "Inescapable Blaze",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Targets:         TargetAny(),
		Purpose:         game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 6)), Answers: game.AnswerRemove},
		OnResolve:       damageToFirstTarget(6),
	})
}
