package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scorn Effigy — Artifact Creature — Scarecrow {3}, 2/3 (Kaldheim, #5505):
//
//	"Foretell {0}"
//
// Foretell is the shared special action (foretell.go, CR 702.143): pay
// {2} on your turn to exile it face down, then cast it on a later turn
// for nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "9ef2a604-4053-4ce0-bc08-43d9cc0b15cc",
		Name:           "Scorn Effigy",
		Completeness:   effects.CompletenessFull,
		SpecialActions: []game.SpecialAction{effects.Foretell("{0}")},
	})
}
