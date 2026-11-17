package effects

// Timberland Ancient — Creature — Treefolk {4}{G}{G}, 6/5 (March of the
// Machine):
//
//	"Reach, trample
//	 Forestcycling {2} ({2}, Discard this card: Search your library for
//	 a Forest card, reveal it, put it into your hand, then shuffle.)"
//
// Forestcycling is the shared Typecycling constructor (CR 702.29e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e71eca6f-388c-4750-b70e-3ee1832931f0",
		Name:            "Timberland Ancient",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Activated: []ActivatedAbility{
			Typecycling("Forestcycling", "{2}", "a Forest card", IsLandWithSubtype("Forest")),
		},
	})
}
