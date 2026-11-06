package effects

// Monoist Sentry — Artifact Creature — Robot {B}, 4/1 (Edge of
// Eternities, #4461):
//
//	"Defender"
//
// A vanilla body with one printed keyword: a creature with defender
// can't attack (CR 702.3b). It can still block, and an effect that lets
// a creature attack as though it didn't have defender lets it attack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0b4cb111-ecf7-4ff7-a9e3-c944054f3370",
		Name:            "Monoist Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
