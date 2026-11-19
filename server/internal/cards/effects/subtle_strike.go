package effects

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
	Register(Spec{
		OracleID:     "08122109-6287-4d78-9242-2da8a25022b0",
		Name:         "Subtle Strike",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeWithPurpose(ModeDoing("Target creature gets -1/-1 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(-1, -1, "Subtle Strike")), ForTargets(HarmUntilEOT(0, -1, -1))),
			ModeDoing("Put a +1/+1 counter on target creature.",
				TargetCreature("target creature"), PlusOneCounterOnTheModesTarget),
		),
	})
}
