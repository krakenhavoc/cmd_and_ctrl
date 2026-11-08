package effects

// Shepherding Spirits — Creature — Spirit {4}{W}{W}, 4/5 (Duskmourn:
// House of Horror, #4466):
//
//	"Flying
//	 Plainscycling {2} ({2}, Discard this card: Search your library for
//	 a Plains card, reveal it, put it into your hand, then shuffle.)"
//
// Plainscycling is the shared Typecycling constructor (CR 702.29e), as on
// Daggermaw Megalodon: any card with the Plains subtype, basic or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e8ff3e99-eca9-44a5-baac-99ee34df3c36",
		Name:            "Shepherding Spirits",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			Typecycling("Plainscycling", "{2}", "a Plains card", IsLandWithSubtype("Plains")),
		},
	})
}
