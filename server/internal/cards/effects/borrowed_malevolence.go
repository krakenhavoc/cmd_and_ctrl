package effects

// Borrowed Malevolence — Instant {B}:
//
//	"Escalate {2} (Pay this cost for each mode chosen beyond the
//	 first.)
//	 Choose one or both —
//	 • Target creature gets +1/+1 until end of turn.
//	 • Target creature gets -1/-1 until end of turn."
//
// Escalate (CR 702.120a, #2126): {2} for the second mode. Each bullet
// has its own target, and the two may be the same creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee6e0279-9606-4877-a3a0-5d9dd557ff15",
		Name:         "Borrowed Malevolence",
		Completeness: CompletenessFull,
		Modes: Escalating(ChooseN("Choose one or both", 1, 2,
			ModeDoing("Target creature gets +1/+1 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(1, 1, "Borrowed Malevolence")),
			ModeDoing("Target creature gets -1/-1 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(-1, -1, "Borrowed Malevolence")),
		), EscalateMana("{2}")),
	})
}
