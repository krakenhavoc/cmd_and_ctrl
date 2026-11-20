package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kuro's Taken — Creature — Rat Samurai {1}{B}, 1/1 (Saviors of Kamigawa,
// #5601):
//
//	"Bushido 1
//	 {1}{B}: Regenerate this creature."
//
// Bushido 1 is the shared keyword trigger (CR 702.45a); the second row is
// the shared regenerateThisRow, one regeneration shield (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b787ba5-0b87-449f-98cd-9454445e2f48",
		Name:         "Kuro's Taken",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Bushido(1, "Kuro's Taken")},
		Activated:    []ActivatedAbility{regenerateThisRow("{1}{B}")},
	})
}
