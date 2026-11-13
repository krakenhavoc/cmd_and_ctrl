package effects

// Aura Extraction — Instant {1}{W} (Onslaught):
//
//	"Put target enchantment on top of its owner's library.
//	 Cycling {2}"
//
// Set Adrift's tuck (putTargetOnTopOfOwnersLibrary) over an
// enchantment, any player's: it goes on top of its OWNER's library,
// whoever controls it, and a commander gets the CR 903.9 offer on the
// way. A token tucked this way ceases to exist (CR 111.8). Cycling is the
// keyword's own hand ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84100106-1f97-4859-aadd-a35796ee000f",
		Name:         "Aura Extraction",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve:    putTargetOnTopOfOwnersLibrary,
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
