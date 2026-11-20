package effects

// Chaplain's Blessing — Sorcery {W} (Shadows over Innistrad, #5543):
//
//	"You gain 5 life."
//
// Sacred Nectar's life gain at five.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "195837b4-e5e4-479d-8dab-6ac77f9fcd83",
		Name:         "Chaplain's Blessing",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(GainLife{Amount: 5}),
	})
}
