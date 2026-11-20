package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tocatli Honor Guard — {1}{W} Creature — Human Soldier 1/3:
//
//	"Creatures entering don't cause abilities to trigger."
//
// Torpor Orb on a 1/3 body (#1735).
func init() {
	effects.Register(effects.Spec{
		OracleID:           "ac946047-4933-4b55-a4eb-85abde1839bb",
		Name:               "Tocatli Honor Guard",
		Completeness:       effects.CompletenessFull,
		TriggerSuppressors: []game.TriggerSuppressor{effects.CreaturesEnteringDontTrigger("Tocatli Honor Guard")},
	})
}
