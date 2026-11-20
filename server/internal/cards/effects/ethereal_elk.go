package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ethereal Elk — Creature — Elk Spirit {3}{G}{G}, 3/3 (Core Set 2020, #5518):
//
//	"Trample
//	 When this creature enters, you may search your library and/or
//	 graveyard for a card named Vivien, Nature's Avenger, reveal it, and put
//	 it into your hand. If you search your library this way, shuffle."
//
// The War of the Spark partner-search trigger
// (mayFetchNamedFromLibraryOrGraveyard): the zones are asked first, and
// the library is shown and shuffled only if it was searched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "82d643ef-8380-4fdc-a88d-ed205324dbd9",
		Name:            "Ethereal Elk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{mayFetchNamedFromLibraryOrGraveyard("Ethereal Elk", "Vivien, Nature's Avenger")},
	})
}
