package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subtle Strike — Instant {1}{B} (Foundations Jumpstart, #5929):
//
//	"Choose one or both —
//	 • Target creature gets -1/-1 until end of turn.
//	 • Put a +1/+1 counter on target creature."
//
// Azula Always Lies prints the same two bullets, so the modes are
// built by the same shape: each bullet has its own target, and a
// player who chooses both may name two different creatures.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08122109-6287-4d78-9242-2da8a25022b0",
		Name:         "Subtle Strike",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -1/-1 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(-1, -1, "Subtle Strike")), effects.ForTargets(effects.HarmUntilEOT(0, -1, -1))),
			effects.ModeDoing("Put a +1/+1 counter on target creature.",
				effects.TargetCreature("target creature"), effects.PlusOneCounterOnTheModesTarget),
		),
	})
}
