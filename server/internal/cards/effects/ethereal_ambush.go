package effects

// Ethereal Ambush — Instant {3}{G}{U} (Fate Reforged,
// #5552):
//
//	"Manifest the top two cards of your library. (To manifest a card, put
//	 it onto the battlefield face down as a 2/2 creature. Turn it face up
//	 any time for its mana cost if it's a creature card.)"
//
// Soul Summons for two: CR 701.40e makes each card its own manifest, so the
// second re-reads the top after the first has entered. Nothing is revealed,
// and an empty library manifests nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0604afb-eae2-41c8-b390-dbc0f7e025ed",
		Name:         "Ethereal Ambush",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(Manifest{N: 2}),
	})
}
