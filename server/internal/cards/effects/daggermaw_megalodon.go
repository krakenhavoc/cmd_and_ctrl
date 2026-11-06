package effects

// Daggermaw Megalodon — Creature — Shark {4}{U}{U}, 5/7 (Duskmourn:
// House of Horror, #4466):
//
//	"Vigilance
//	 Islandcycling {2} ({2}, Discard this card: Search your library for
//	 an Island card, reveal it, put it into your hand, then shuffle.)"
//
// Islandcycling is the shared Typecycling constructor (CR 702.29e), as on
// Giant Koi: any card with the Island subtype, basic or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dc638413-c4f8-4f6a-a003-5ce8a10a8f94",
		Name:            "Daggermaw Megalodon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{
			Typecycling("Islandcycling", "{2}", "an Island card", IsLandWithSubtype("Island")),
		},
	})
}
