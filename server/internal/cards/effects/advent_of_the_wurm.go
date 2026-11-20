package effects

// Advent of the Wurm — Instant {1}{G}{G}{W}:
//
//	"Create a 5/5 green Wurm creature token with trample."
//
// One token from the shared table, the 5/5 green Wurm with trample.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb62aa4b-c11b-4195-ae85-cff8f78ce17b",
		Name:         "Advent of the Wurm",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("5/5 green Wurm with trample"), N: 1},
		),
	})
}
