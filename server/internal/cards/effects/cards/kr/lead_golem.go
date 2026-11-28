package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lead Golem — Artifact Creature — Golem {5}, 3/5 (Mirage, #5633):
//
//	"Whenever this creature attacks, it doesn't untap during its
//	 controller's next untap step."
//
// Apes of Rath's trigger on a Golem (OldB813DoesntUntapWhenThisAttacks).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5810c1e2-5a36-4a87-be4e-890ba1627b46",
		Name:         "Lead Golem",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.OldB813DoesntUntapWhenThisAttacks("Lead Golem")},
	})
}
