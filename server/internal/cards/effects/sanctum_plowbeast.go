package effects

// Sanctum Plowbeast — Artifact Creature — Beast {4}{W}{U}, 3/6 (Alara
// Reborn, #5586):
//
//	"Defender
//	 Plainscycling {2}, islandcycling {2} ({2}, Discard this card: Search
//	 your library for a Plains or Island card, reveal it, put it into your
//	 hand, then shuffle.)"
//
// Defender is the printed keyword. The two landcycling keywords are two
// cycling abilities (CR 702.29e), each the shared Typecycling row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff3fae6e-d8a1-46b3-a6cc-fb4e8800ee68",
		Name:            "Sanctum Plowbeast",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{
			Typecycling("Plainscycling", "{2}", "a Plains card", IsLandWithSubtype("Plains")),
			Typecycling("Islandcycling", "{2}", "an Island card", IsLandWithSubtype("Island")),
		},
	})
}
