package effects

// Urza's Rebuff — Instant for {1}{U}{U} (The Brothers' War):
//
//	"Choose one —
//	• Counter target spell.
//	• Tap up to two target creatures."
func init() {
	Register(Spec{
		OracleID:     "db3b294a-87b1-4dda-b1f6-ea63a9722faf",
		Name:         "Urza's Rebuff",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Counter target spell.",
				TargetSpell("target spell"),
				CounterTheModesTarget),
			ModeDoing("Tap up to two target creatures.",
				TargetCreature("up to two target creatures").WithCount(0, 2),
				tapTheModesTargets),
		),
	})
}
