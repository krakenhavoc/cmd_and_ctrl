package effects

// Pale Recluse — Creature — Spider {4}{G}{W}, 4/5 (Alara Reborn, #5586):
//
//	"Reach (This creature can block creatures with flying.)
//	 Forestcycling {2}, plainscycling {2} ({2}, Discard this card: Search
//	 your library for a Forest or Plains card, reveal it, put it into your
//	 hand, then shuffle.)"
//
// Reach is the printed keyword. The two landcycling keywords are two
// cycling abilities (CR 702.29e), each the shared Typecycling row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "488ca652-653b-4bb3-891c-874efdd418ff",
		Name:            "Pale Recluse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{
			Typecycling("Forestcycling", "{2}", "a Forest card", IsLandWithSubtype("Forest")),
			Typecycling("Plainscycling", "{2}", "a Plains card", IsLandWithSubtype("Plains")),
		},
	})
}
