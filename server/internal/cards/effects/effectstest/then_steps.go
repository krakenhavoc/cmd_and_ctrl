package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// then_steps_test.go — #4162 and #4163 (CR 608.2c): a continuation's
// later instruction waits for an earlier one's prompt, and a later step
// reads what an earlier one recorded.
//
// The shapes, each with a real card where one exists:
//
//   - a prompt's Then written as continuation steps, looping through its
//     own record (Seize the Spotlight's fortunes), on a dredge offer;
//   - a *ThenForEffect callback's continuation (Scavenging Ooze's exile
//     callback), on a CR 616 counter order;
//   - a spell's ResolveSteps whose later step reads ids from the record
//     (Martial Coup), on a CR 616 token order;
//   - a spell's ResolveSteps whose later step reads a count from the
//     record (Cut a Deal), on a dredge offer, both answers;
//   - a continuation parked above the item's own Do walk, whose later
//     step runs after the continuation, live and across a strict restore
//     taken mid-pause, with the record restored (the probe below: no
//     catalog continuation's first step asks a prompt that is data).
//
// A dredge offer and a CR 616 order resume through closures, so a table
// holding one is not a restore point; an undo (a clone) stands in for the
// restore there, as in resolution_order_batch1_test.go. A scry is data.

const ThenStepsProbeOracle = "test-4162-then-steps-probe"

const thenStepsProbeRecord = "test-then-steps/n"

// thenStepsProbeLook is the hand size the probe's Effect saw as its
// last step ran (-1: not run).
var ThenStepsProbeLook = -1

// thenStepsProbeRest is "scry 1, then draw the number chosen".
var thenStepsProbeRest = effects.RegisterThenSteps("test/then-steps-probe",
	effects.Scry{N: 1},
	effects.Instruction(func(ctx *effects.Context) error {
		return effects.DrawCards{N: ctx.StepRecord(thenStepsProbeRecord).N()}.Apply(ctx)
	}),
)

// thenStepsProbeChosen is the number's keyed branch.
var thenStepsProbeChosen = effects.NumberThen("test/then-steps-probe-chosen",
	func(ctx *effects.Context, r game.NumberAnswer) error {
		ctx.RecordStep(thenStepsProbeRecord, nil, r.Number)
		return thenStepsProbeRest.Run(ctx)
	})

func init() {
	effects.Register(effects.Spec{
		OracleID: ThenStepsProbeOracle,
		Name:     "Then Steps Probe",
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Choose a number. Scry 1, then draw that many cards. Then look at your hand.",
			Cost:    effects.TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: effects.Do(
				effects.ChooseNumber{
					Min: 0, Max: 2, Question: "Then Steps Probe — how many?", Unit: game.PayAmountCards,
					Then: thenStepsProbeChosen,
				},
				effects.Instruction(func(ctx *effects.Context) error {
					ThenStepsProbeLook = ctx.Game.PlayerByIDForEffect(ctx.Controller()).Hand.Size()
					return nil
				}),
			),
		}},
	})
}
