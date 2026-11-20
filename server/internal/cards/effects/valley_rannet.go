package effects

// Valley Rannet — Creature — Beast {4}{R}{G}, 6/3 (Alara Reborn, #5586):
//
//	"Mountaincycling {2}, forestcycling {2} ({2}, Discard this card:
//	 Search your library for a Mountain or Forest card, reveal it, put it
//	 into your hand, then shuffle.)"
//
// The two landcycling keywords are two cycling abilities (CR 702.29e),
// each the shared Typecycling row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0a3b261-3daf-4082-8bc5-e66326a6c382",
		Name:         "Valley Rannet",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			Typecycling("Mountaincycling", "{2}", "a Mountain card", IsLandWithSubtype("Mountain")),
			Typecycling("Forestcycling", "{2}", "a Forest card", IsLandWithSubtype("Forest")),
		},
	})
}
