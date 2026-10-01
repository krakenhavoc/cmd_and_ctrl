package effects

// Call In a Professional — Instant {2}{R}:
//
//	"Players can't gain life this turn. Damage can't be prevented this
//	 turn. Call In a Professional deals 3 damage to any target. (Shield
//	 counters don't prevent this damage as they're removed.)"
//
// Skullcrack's three sentences with "any target" (ADR 0107 §5). The
// reminder text is about shield counters, whose "remove a shield
// counter instead" is a replacement, not a prevention effect, so it is
// not stopped by "can't be prevented" — the engine gives a shield
// counter no behaviour of its own, so there is nothing to apply.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "ca54e14c-67ee-4de2-bc0f-19a5f9461d37",
		Name:         "Call In a Professional",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve:    skullcrackShape(3),
	})
}
