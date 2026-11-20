package effects

// Jhessian Zombies — Creature — Zombie {4}{U}{B}, 2/4 (Alara Reborn,
// #5586):
//
//	"Fear (This creature can't be blocked except by artifact creatures
//	 and/or black creatures.)
//	 Islandcycling {2}, swampcycling {2} ({2}, Discard this card: Search
//	 your library for an Island or Swamp card, reveal it, put it into your
//	 hand, then shuffle.)"
//
// Fear is the printed keyword (CR 702.36). The two landcycling keywords are
// two cycling abilities (CR 702.29e), each the shared Typecycling row
// (Giant Koi's islandcycling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "40cc8ac6-114f-439d-816d-f65cd52bbbd2",
		Name:            "Jhessian Zombies",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"fear"},
		Activated: []ActivatedAbility{
			Typecycling("Islandcycling", "{2}", "an Island card", IsLandWithSubtype("Island")),
			Typecycling("Swampcycling", "{2}", "a Swamp card", IsLandWithSubtype("Swamp")),
		},
	})
}
