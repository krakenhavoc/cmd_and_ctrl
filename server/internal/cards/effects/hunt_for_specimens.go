package effects

// Hunt for Specimens — Sorcery {1}{B} (EDHREC rank 18326):
//
//	"Create a 1/1 black and green Pest creature token with "When this
//	 token dies, you gain 1 life."
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// The Pest is the same `token:pest` catalog template Beledros
// Witherbloom and Sedgemoor Witch make, so its life-gain trigger is
// printed on it. Learn follows the token (CR 701.48a, learn.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53f1bcff-55fb-4d38-b108-dcbed623f488",
		Name:         "Hunt for Specimens",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(CreateToken{Template: PestToken(), N: 1}, Learn{}),
	})
}
