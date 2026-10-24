package effects

// Cram Session — Sorcery {1}{B/G} (EDHREC rank 25366):
//
//	"You gain 4 life.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// Learn is the last instruction: a Learn step after the life gain
// (CR 701.48a, learn.go). The cost is hybrid, which is the deck
// importer's concern and not the card's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6e5b8bb2-9dd8-4986-9d0a-659114be5a2a",
		Name:         "Cram Session",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(GainLife{Amount: 4}, Learn{}),
	})
}
