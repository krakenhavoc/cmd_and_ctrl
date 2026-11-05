package effects

// Azula Always Lies — Instant — Lesson {1}{B}:
//
//	"Choose one or both —
//	 • Target creature gets -1/-1 until end of turn.
//	 • Put a +1/+1 counter on target creature."
//
// Borrowed Malevolence's "one or both" with a counter for its second
// mode. Each mode has its own target, so both modes may name the same
// creature (CR 115.3): a -1/-1 and a +1/+1 counter on one creature. A
// mode whose target is illegal as the spell resolves does nothing, and
// the other still happens (CR 608.2b).
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "f56c6c35-d09b-4bee-a685-98e9afed13d1",
		Name:         "Azula Always Lies",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeWithPurpose(ModeDoing("Target creature gets -1/-1 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(-1, -1, "Azula Always Lies")), ForTargets(HarmUntilEOT(0, -1, -1))),
			ModeDoing("Put a +1/+1 counter on target creature.",
				TargetCreature("target creature"), PlusOneCounterOnTheModesTarget),
		),
	})
}
