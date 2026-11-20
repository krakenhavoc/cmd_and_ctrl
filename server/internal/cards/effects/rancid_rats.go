package effects

// Rancid Rats — Creature — Zombie Rat {1}{B}, 1/1
// (Shadows over Innistrad, #5543):
//
//	"Skulk (This creature can't be blocked by creatures with greater
//	 power.)
//	 Deathtouch (Any amount of damage this deals to a creature is enough
//	 to destroy it.)"
//
// Skulk and deathtouch are engine keywords carried as printed data (CR
// 702.118, CR 702.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e3552730-fe09-443f-b432-9640b69c7bbb",
		Name:            "Rancid Rats",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"skulk", "deathtouch"},
	})
}
