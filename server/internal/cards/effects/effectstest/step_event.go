package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// step_event_test.go — #588: the generic step announcement is a
// trigger event. A probe declares "at the beginning of combat on your
// turn"; it fires on the controller's begin-combat step, goes on the
// stack, and stays quiet on an opponent's.

const StepProbeOracle = "test-step-event-probe"

func init() {
	effects.Register(effects.Spec{
		OracleID: StepProbeOracle,
		Name:     "Step Probe",
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Step Probe — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
