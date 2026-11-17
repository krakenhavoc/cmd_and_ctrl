package effects

// Shardless Outlander — Artifact Creature — Construct Scout {7}, 6/5
// (Foundations Jumpstart, #5464):
//
//	"Trample (This creature can deal excess combat damage to the player
//	 or planeswalker it's attacking.)
//	 Basic landcycling {2} ({2}, Discard this card: Search your library
//	 for a basic land card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// The printed keyword and the shared basic landcycling row (cycling.go),
// as Hexhaven Battalion has it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "823afa57-faac-4a43-a403-c1bb0446d594",
		Name:            "Shardless Outlander",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
