package effects

// Furnace Host Charger — Creature — Phyrexian Giant {5}{R}, 5/5 (March of
// the Machine):
//
//	"Haste
//	 Mountaincycling {2} ({2}, Discard this card: Search your library for
//	 a Mountain card, reveal it, put it into your hand, then shuffle.)"
//
// Mountaincycling is the shared Typecycling constructor (CR 702.29e), as on
// Chartooth Cougar.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fea60c45-cc36-4dd1-8475-23a9361e370d",
		Name:            "Furnace Host Charger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{
			Typecycling("Mountaincycling", "{2}", "a Mountain card", IsLandWithSubtype("Mountain")),
		},
	})
}
