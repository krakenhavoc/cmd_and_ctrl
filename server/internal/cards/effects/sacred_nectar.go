package effects

// Sacred Nectar — Sorcery {1}{W} (Portal, #5630):
//
//	"You gain 4 life."
//
// One life-gain step for its controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "30870ee5-6ad7-48a9-983e-d3b018f2344f",
		Name:         "Sacred Nectar",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(GainLife{Amount: 4}),
	})
}
