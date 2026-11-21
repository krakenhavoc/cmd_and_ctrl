package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const RunAtProbeOracle = "test-4155-run-at-probe"

// runAtProbeRan is the players the probe's last step ran for, in order.
var RunAtProbeRan []uuid.UUID

// runAtProbeChosen is the number's keyed branch: it does nothing.
var runAtProbeChosen = effects.NumberThen("test/run-at-probe-chosen",
	func(*effects.Context, game.NumberAnswer) error { return nil })

// runAtProbeSteps is "choose a number, then note which player you are": the
// first step asks the player the occurrence names, the second runs after the
// answer and reads the occurrence again.
var runAtProbeSteps = effects.RegisterThenSteps("test/run-at-probe",
	effects.Instruction(func(ctx *effects.Context) error {
		return effects.ChooseNumber{
			Chooser: effects.SeatAtOccurrence(ctx), Min: 0, Max: 1, Question: "Run At Probe — how many?", Unit: game.PayAmountCards,
			Then: runAtProbeChosen,
		}.Apply(ctx)
	}),
	effects.Instruction(func(ctx *effects.Context) error {
		RunAtProbeRan = append(RunAtProbeRan, effects.SeatAtOccurrence(ctx))
		return nil
	}),
)

// runAtProbeGo is a player's yes, keyed (#4253 retired MayChoice's OnYes):
// the player rides the prompt as its Carry.
var runAtProbeGo = effects.MayYes("test/4253-run-at-probe-go", func(c *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	return runAtProbeSteps.RunAt(c, effects.SeatIndexOf(c.Game, r.Carry[0]))
})

func init() {
	effects.Register(effects.Spec{
		OracleID: RunAtProbeOracle,
		Name:     "Run At Probe",
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Ask two other players, side by side.",
			Cost:    effects.TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				for _, p := range ctx.Opponents()[:2] {
					player := p
					if err := (effects.MayChoice{
						Player:   player,
						Question: "Run At Probe — go?",
						Then:     runAtProbeGo,
						Carry:    []uuid.UUID{player},
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})),
		}},
	})
}
