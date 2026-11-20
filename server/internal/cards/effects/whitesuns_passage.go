package effects

// Whitesun's Passage — Instant {1}{W} (Scars of Mirrodin, #5577):
//
//	"You gain 5 life."
//
// Sacred Nectar's shape at 5.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df702285-3577-47cc-aa1e-868921cdf6f2",
		Name:         "Whitesun's Passage",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(GainLife{Amount: 5}),
	})
}
