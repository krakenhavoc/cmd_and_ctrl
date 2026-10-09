package effects

// Undulating Witness — Creature — Serpent {4}{U}, 3/5:
//
//	"Flying
//	 {2}: This creature gets +1/-1 until end of turn.
//	 Basic landcycling {2}"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "77b684b8-227b-40de-80b4-15d0e70914f4",
		Name:            "Undulating Witness",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			{
				Label:  "{2}: This creature gets +1/-1 until end of turn.",
				Cost:   ManaCost("{2}"),
				Effect: thisCreatureUntilEOT("Undulating Witness — +1/-1", 1, -1),
			},
			BasicLandcycling("{2}"),
		},
	})
}
