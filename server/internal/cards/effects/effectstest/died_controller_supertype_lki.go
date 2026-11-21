package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- every switched site --------------------------------------------

// The probe: a battlefield permanent whose LTB ability runs the hook
// during the harvest — while the CR 603.10 counter snapshot a few of
// the conditions read (Resourceful Defense, The Ozolith) still exists
// — and never triggers.
const LtbControllerProbeOracle = "test-1682-ltb-controller-probe"

var LtbControllerProbeHook func(ev game.Event, g *game.Game)

func init() {
	effects.Register(effects.Spec{
		OracleID: LtbControllerProbeOracle,
		Name:     "LKI Probe",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if LtbControllerProbeHook != nil {
					LtbControllerProbeHook(ev, g)
				}
				return false
			},
			Key:    "LKI Probe — never",
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
}
