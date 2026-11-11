package effects

// Timeless Dragon — Creature — Dragon {3}{W}{W}, 5/5:
//
//	"Flying
//	 Plainscycling {2}
//	 Eternalize {2}{W}{W}"
//
// Flying is a printed keyword, Plainscycling is the shared Typecycling
// constructor (a Plains card, not only a basic one) and Eternalize is the
// shared graveyard keyword (CR 702.129a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a5ae93e9-0385-47fd-9d8f-306e9a6b7fe6",
		Name:            "Timeless Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			Typecycling("Plainscycling", "{2}", "a Plains card", IsLandWithSubtype("Plains")),
			Eternalize("{2}{W}{W}"),
		},
	})
}
