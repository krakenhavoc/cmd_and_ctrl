package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// triggers_common_test.go — the constructor vocabulary (#579). The 400+
// migrated card files are the broad guard; these pin the contract the
// constructors promise on their own: the item goes on the stack, Do
// defaults the controller, Optional asks first, and the predicates
// read the event the way their names say.

const TriggerProbeOracle = "test-triggers-common-probe"

func init() {
	effects.Register(effects.Spec{
		OracleID: TriggerProbeOracle,
		Name:     "Constructor Probe",
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Constructor Probe — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
			effects.Optional(effects.WhenThisDies("Constructor Probe — draw a card", effects.Do(effects.DrawCards{N: 1})),
				"Constructor Probe — draw a card?"),
		},
	})
}
