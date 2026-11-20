package effects

// Force of Savagery — Creature — Elemental {2}{G}, 8/0 (Future Sight, #5594):
//
//	"Trample"
//
// A printed keyword and nothing else. Its toughness of 0 is the card's,
// read off Scryfall: with nothing raising it, it is put into its owner's
// graveyard by the state-based check as it enters (CR 704.5f), which is
// what the printed card does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "26fbfab5-f479-49f2-b36b-a3ebf964741f",
		Name:            "Force of Savagery",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
