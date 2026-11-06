package effects

// Saber-Tooth Moose-Lion — Creature — Elk Cat {4}{G}{G}, 7/7:
//
//	"Reach
//	 Forestcycling {2} ({2}, Discard this card: Search your library for a
//	 Forest card, reveal it, put it into your hand, then shuffle.)"
//
// Reach is printed. Forestcycling is typecycling (CR 702.29e) for any
// card with the Forest subtype, Rocksteady, Crash Courser's line.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "7af493e3-15a1-42d6-a66c-82aacabe816f",
		Name:            "Saber-Tooth Moose-Lion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{
			Typecycling("Forestcycling", "{2}", "a Forest card", IsLandWithSubtype("Forest")),
		},
	})
}
