package effects

// Thoughtbind — Instant {2}{U} (Champions of Kamigawa, #5603):
//
//	"Counter target spell with mana value 4 or less."
//
// Disdainful Stroke turned around. On the stack an {X} counts as the
// number chosen for it (CR 202.3e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f31bb2b4-061c-4ba6-9f91-374f40fe6b05",
		Name:         "Thoughtbind",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell with mana value 4 or less", ManaValueLE(4)),
		OnResolve:    counterTheTargetSpell,
	})
}
