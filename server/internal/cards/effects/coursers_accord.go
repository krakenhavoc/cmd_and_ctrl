package effects

// Coursers' Accord — Sorcery {4}{G}{W}:
//
//	"Create a 3/3 green Centaur creature token, then populate.
//	 (Create a token that's a copy of a creature token you control.)"
//
// The token is made first, so it is a candidate for the populate that
// follows (and the only one when you control no other creature token).
// Populate is the shared primitive in populate.go.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23674648-d1a0-437d-8a0b-4173c75b4507",
		Name:         "Coursers' Accord",
		Completeness: CompletenessFull,
		// Populate waits for a prompt the Centaur's creation asks (a CR 616 order; CR 608.2c,
		// #4155) and may copy the token it made.
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("3/3 green Centaur"), N: 1},
			Populate{},
		),
	})
}
