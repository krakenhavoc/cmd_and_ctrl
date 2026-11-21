package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Probes for the two shapes no catalog card writes yet: an If inside a
// spell's ResolveSteps, and a bullet's steps followed by a later bullet.
// Plus #4070's modal spell: the first bullet's one damage instance hits
// Khârn and a player under Selfless Squire's shield, and the second
// bullet looks at the Squire's triggers.
const (
	ResolveStepsProbeOracle = "test-4019-resolve-steps-probe"
	ModeStepsProbeOracle    = "test-4019-mode-steps-probe"
	ModeSquireProbeOracle   = "test-4070-mode-squire-probe"
)

// modeStepsSawHand records the hand size as the probe's second bullet
// begins (-1: not run).
var ModeStepsSawHand = -1

func init() {
	effects.Register(effects.Spec{
		OracleID:  ResolveStepsProbeOracle,
		Name:      "Resolve Steps Probe",
		OnResolve: effects.ResolveSteps(effects.If{Cond: doWalkAlways{}, Then: []effects.Applier{effects.Scry{N: 1}, effects.DrawCards{N: 1}}}, effects.GainLife{Amount: 1}),
	})
	effects.Register(effects.Spec{
		OracleID: ModeStepsProbeOracle,
		Name:     "Mode Steps Probe",
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeSteps("Scry 1, then draw a card.", nil, effects.Scry{N: 1}, effects.DrawCards{N: 1}),
			effects.ModeSteps("Look at your hand.", nil, effects.Instruction(func(ctx *effects.Context) error {
				for _, p := range ctx.Game.Seats {
					if p.ID == ctx.Controller() {
						ModeStepsSawHand = p.Hand.Size()
					}
				}
				return nil
			})),
		),
	})
	effects.Register(effects.Spec{
		OracleID: ModeSquireProbeOracle,
		Name:     "Mode Squire Probe",
		Modes: effects.ChooseN("Choose two", 2, 2,
			effects.ModeDoing("1 damage to target creature and 2 damage to you.", effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, _ int) error { return doWalkProbeKharnAndYou{}.Apply(ctx) }),
			effects.ModeDoing("Look at the Squire.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error { return doWalkProbeSeeSquire{}.Apply(ctx) }),
		),
	})
}
