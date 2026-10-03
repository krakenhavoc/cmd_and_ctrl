package effects

// Shielded Passage — Instant {W}:
//
//	"Prevent all damage that would be dealt to target creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): Indestructible Aura's text, the
// same not-one-use shield pinned to the target.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a6933789-bb4f-4d45-975e-402b8e0b0e68",
		Name:         "Shielded Passage",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    shieldTargetSpell(),
	})
}
