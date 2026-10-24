package effects

// Step Right Up — Sorcery {3}{B}:
//
//	"Open two Attractions. (Put the top two cards of your Attraction
//	 deck onto the battlefield.)"
//
// Two opens, one after the other (CR 701.51b): each is its own entry and
// its own "whenever you open an Attraction" (CR 701.51c). With one card
// left in the Attraction deck, the second opens nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "aaf6b3ed-2396-46d7-a32d-7d24cc2eb449",
		Name:         "Step Right Up",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(OpenAttractions{N: 2}),
	})
}
