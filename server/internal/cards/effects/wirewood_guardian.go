package effects

// Wirewood Guardian — Creature — Elf Mutant {5}{G}{G}, 6/6 (Scourge):
//
//	"Forestcycling {2} ({2}, Discard this card: Search your library for a
//	 Forest card, reveal it, put it into your hand, then shuffle.)"
//
// The shared typecycling ability (CR 702.29e) on a vanilla body. The
// search finds any card with the Forest subtype, a nonbasic Forest
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61ade140-5b11-4882-82e5-a075fc80b260",
		Name:         "Wirewood Guardian",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			Typecycling("Forestcycling", "{2}", "a Forest card", IsLandWithSubtype("Forest")),
		},
	})
}
