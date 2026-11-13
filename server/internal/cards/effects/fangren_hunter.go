package effects

// Fangren Hunter — Creature — Beast {3}{G}{G}, 4/4 (Mirrodin, #5606):
//
//	"Trample"
//
// Trample is the printed keyword the combat-damage assignment reads
// (CR 702.19b); the cost, the type and the 4/4 ride the card from
// Scryfall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c5dc5546-e9e5-4b5b-b812-5716d4bdee0e",
		Name:            "Fangren Hunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
