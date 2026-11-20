package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vengeful Reaper — Creature — Angel Cleric {3}{B}, 2/3 (Kaldheim, #5505):
//
//	"Flying, deathtouch, haste
//	 Foretell {1}{B}"
//
// The keywords ride PrintedKeywords; foretell is the shared special
// action (foretell.go, CR 702.143), declared at the printed cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d776be0-cece-4496-a3b0-b7424eca77fd",
		Name:            "Vengeful Reaper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch", "haste"},
		SpecialActions:  []game.SpecialAction{Foretell("{1}{B}")},
	})
}
