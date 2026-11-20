package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doomskar — Sorcery {3}{W}{W} (Kaldheim, #5505):
//
//	"Destroy all creatures.
//	 Foretell {1}{W}{W}"
//
// Day of Judgment's sweep (wrathDestroyAllCreatures) with the shared
// foretell special action (CR 702.143).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "38ece0fa-2c94-4eae-b6ab-74b783862f3e",
		Name:           "Doomskar",
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Foretell("{1}{W}{W}")},
		OnResolve:      wrathDestroyAllCreatures,
	})
}
