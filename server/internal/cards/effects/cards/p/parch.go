package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Parch — Instant {1}{R}:
//
//	"Choose one —
//	 • Parch deals 2 damage to any target.
//	 • Parch deals 4 damage to target blue creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7366374a-db19-4fb0-9c35-a6e466a9bbc3",
		Name:         "Parch",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Parch deals 2 damage to any target.", effects.TargetAny(), effects.SocB20DamageToTheModeTarget(2)),
			effects.ModeDoing("Parch deals 4 damage to target blue creature.",
				effects.TargetCreature("target blue creature", effects.OfColor("U")), effects.SocB20DamageToTheModeTarget(4)),
		),
	})
}
