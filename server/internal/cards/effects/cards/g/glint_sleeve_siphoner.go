package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glint-Sleeve Siphoner — Creature — Human Rogue {1}{B}, 2/1:
//
//	"Menace
//	 Whenever this creature enters or attacks, you get {E} (an energy
//	 counter).
//	 At the beginning of your upkeep, you may pay {E}{E}. If you do, you
//	 draw a card and you lose 1 life."
//
// ADR 0129 §3 (#1995): the energy is paid as the upkeep trigger
// resolves (CR 118.12); the prompt holds the upkeep until it is
// answered, so the draw comes before the draw step.
//
// No simplification.
// glintSleeveSiphonerDrawAndLose is the draw and then the life loss: the loss
// waits for a prompt the draw asks (a dredge offer; CR 608.2c, #4155).
var glintSleeveSiphonerDrawAndLose = effects.RegisterThenSteps("then-steps/glint-sleeve-siphoner-draw-and-lose",
	effects.DrawCards{N: 1},
	effects.Instruction(func(ctx *effects.Context) error {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
	}),
)

var glintSleeveSiphonerPaid = effects.EnergyPaid("pay-unless/glint-sleeve-siphoner-draw-and-lose",
	func(g *game.Game, item *game.StackItem) error {
		return glintSleeveSiphonerDrawAndLose.Run(effects.NewContext(g, item))
	})

func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea34ec71-2f3b-4119-937d-4629cbf7ac5e",
		Name:            "Glint-Sleeve Siphoner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Purpose:         game.Purpose{Energy: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Glint-Sleeve Siphoner — you get {E}", effects.Do(effects.GetEnergy{N: 1})),
			effects.AtYourUpkeep("Glint-Sleeve Siphoner — you may pay {E}{E}",
				effects.MayPayEnergyThen("Glint-Sleeve Siphoner", 2, "draw a card and lose 1 life", glintSleeveSiphonerPaid)),
		},
	})
}
