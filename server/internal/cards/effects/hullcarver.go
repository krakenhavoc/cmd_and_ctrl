package effects

// Hullcarver — Artifact Creature — Robot Assassin {B}, 1/1 (Edge of
// Eternities, #4461):
//
//	"Deathtouch"
//
// A vanilla body with one printed keyword: any damage it deals to a
// creature is enough to destroy that creature (CR 702.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "931c2153-e5e5-4642-bff0-f9dadcc24c9a",
		Name:            "Hullcarver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
