package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harbinger of Night — Creature — Spirit {2}{B}{B}, 2/3 (Mirage):
//
//	"At the beginning of your upkeep, put a -1/-1 counter on each creature."
//
// Soul Snuffers' counter on every creature, its own included, the set read
// before the first counter goes on, as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9827d01d-c822-4f1b-9042-a2092c72c914",
		Name:         "Harbinger of Night",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Harbinger of Night — put a -1/-1 counter on each creature", effects.Cmdr2B25CountersOnEach(false, nil)),
		},
	})
}
