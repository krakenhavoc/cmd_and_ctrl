package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blade Instructor — Creature — Human Soldier {2}{W}, 3/1:
//
//	"Mentor (Whenever this creature attacks, put a +1/+1 counter on
//	 target attacking creature with lesser power.)"
//
// Mentor (CR 702.134) is the shared Mentor() trigger; see
// hammer_dropper.go for what its source-relative clause guarantees.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77cff683-013c-4568-8ae0-f86ef0590fe1",
		Name:         "Blade Instructor",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Mentor("Blade Instructor — mentor")},
	})
}
