package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8b787ba5-0b87-449f-98cd-9454445e2f48",
		Name:         "Kuro's Taken",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Bushido(1, "Kuro's Taken")},
		Activated:    []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{B}")},
	})
}
