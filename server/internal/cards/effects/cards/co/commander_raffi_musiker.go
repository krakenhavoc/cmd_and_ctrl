package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Commander Raffi Musiker — Legendary Creature — Human Officer {2}{B}, 3/2
// (Star Trek):
//
//	When Commander Musiker enters, choose one —
//	• Exile up to one target card from a graveyard. You gain 2 life.
//	• Investigate. You lose 2 life.
//
// The mode is chosen as the trigger goes on the stack (CR 700.2b). Each
// bullet is two instructions, so its second sentence waits for any prompt
// the first one asks (CR 608.2c). The life gain does not depend on a card
// being exiled: "up to one" may be zero.
//
// No simplification.
func init() {
	t := effects.WhenThisEnters("Commander Raffi Musiker — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseOne(
		effects.ModeSteps("Exile up to one target card from a graveyard. You gain 2 life.",
			effects.TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1),
			effects.Instruction(func(ctx *effects.Context) error {
				tr, ok := effects.ModeTarget(ctx, ctx.Occurrence())
				if !ok || tr.Kind != game.TargetCard {
					return nil
				}
				return effects.ExileTarget{Target: tr.ID}.Apply(ctx)
			}),
			effects.GainLife{Amount: 2}),
		effects.ModeSteps("Investigate. You lose 2 life.", nil,
			effects.Investigate{N: 1},
			effects.GainLife{Amount: -2}),
	)
	effects.Register(effects.Spec{
		OracleID:     "4f8a70cb-2487-4e6d-9a3f-c6c5313b8e82",
		Name:         "Commander Raffi Musiker",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
