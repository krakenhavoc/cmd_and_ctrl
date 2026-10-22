package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
var glintSleeveSiphonerDrawAndLose = RegisterThenSteps("then-steps/glint-sleeve-siphoner-draw-and-lose",
	DrawCards{N: 1},
	Instruction(func(ctx *Context) error {
		return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -1)
	}),
)

func init() {
	Register(Spec{
		OracleID:        "ea34ec71-2f3b-4119-937d-4629cbf7ac5e",
		Name:            "Glint-Sleeve Siphoner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Purpose:         game.Purpose{Energy: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Glint-Sleeve Siphoner — you get {E}", Do(GetEnergy{N: 1})),
			AtYourUpkeep("Glint-Sleeve Siphoner — you may pay {E}{E}",
				mayPayEnergyThen("Glint-Sleeve Siphoner", 2, "draw a card and lose 1 life",
					func(g *game.Game, item *game.StackItem) error {
						return glintSleeveSiphonerDrawAndLose.Run(NewContext(g, item))
					})),
		},
	})
}
