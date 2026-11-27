package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Panglacial Wurm — Creature — Wurm {5}{G}{G}, 9/5 (Coldsnap, #5597):
//
//	"Trample
//	 While you're searching your library, you may cast this card from your
//	 library."
//
// The seam is #6216's (game/search_cast.go, CR 701.23a, 608.2g): while its owner
// has an open search of their own library inside a resolution, the card may be
// cast from the library, paid in full, and the search then goes on.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:                  "49d4aef7-d2d1-4f66-bc2a-79ec8cc96565",
		Name:                      "Panglacial Wurm",
		Completeness:              effects.CompletenessFull,
		PrintedKeywords:           []string{"trample"},
		CastWhileSearchingLibrary: true,
	})
}
