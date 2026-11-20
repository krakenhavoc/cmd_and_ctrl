package effects

// Vampire Cutthroat — Creature — Vampire Rogue {B}, 1/1 (Eldritch Moon,
// #5541):
//
//	"Skulk (This creature can't be blocked by creatures with greater power.)
//	 Lifelink (Damage dealt by this creature also causes you to gain that
//	 much life.)"
//
// Both keywords are engine keywords carried as printed data: skulk is
// Furtive Homunculus's (CR 702.118b), lifelink the usual one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2e163679-7303-4e05-9d20-86b76dfa1407",
		Name:            "Vampire Cutthroat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"skulk", "lifelink"},
	})
}
