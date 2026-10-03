package effects

// Indestructible Aura — Instant {W}:
//
//	"Prevent all damage that would be dealt to target creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the recipient family of the
// not-one-use shield. A preventFromSource record pinned to the target
// with no source: every instance of damage to it this turn, from
// anything, is prevented. A creature that leaves the battlefield and
// returns is a new object the shield no longer covers (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e10e8d56-bba6-412d-970e-c24969f32b5b",
		Name:         "Indestructible Aura",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    shieldTargetSpell(),
	})
}
