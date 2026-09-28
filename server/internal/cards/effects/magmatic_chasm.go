package effects

// Magmatic Chasm — Sorcery {1}{R}:
//
//	"Creatures without flying can't block this turn."
//
// Falter at sorcery speed. The set is read live until cleanup (#1650),
// so a creature that arrives later in the turn can't block either; see
// falter.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6cfd2cd8-a86e-48fe-a85a-9a3444b5030c",
		Name:         "Magmatic Chasm",
		Completeness: CompletenessFull,
		OnResolve:    faltering("Magmatic Chasm"),
	})
}
