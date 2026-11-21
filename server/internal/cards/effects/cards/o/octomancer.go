package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Octomancer — Creature — Frog Druid {3}{G}{U}, 3/3:
//
//	"Gift an Octopus (You may promise an opponent a gift as you cast
//	 this spell. If you do, when it enters, they create an 8/8 blue
//	 Octopus creature token.)
//	 At the beginning of each end step, create a token that's a copy of
//	 target creature token that entered the battlefield this turn."
//
// The gift is GiftAnOctopus (CR 702.174i), the entry trigger of a
// permanent (CR 702.174b). The end-step trigger fires at EVERY end step,
// whoever's turn it is, and its target is any creature token that entered
// this turn, whoever controls it — including the Octopus the gift gave an
// opponent. The target is chosen as the trigger goes on the stack (CR
// 603.3d) and is gone-checked at resolution (CR 608.2b); the copy is the
// token's copiable values (CR 707.2) and is made under Octomancer's
// controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1cf7e5a0-cc8f-4bf7-9767-2f3bcae26339",
		Name:         "Octomancer",
		Completeness: effects.CompletenessFull,
		Gift:         effects.GiftAnOctopus(),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtEachStep(game.StepEnd,
					"Octomancer — create a token that's a copy of target creature token that entered this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.CreateTokenCopy{Controller: item.Controller, Copy: id, N: 1}.Apply(ctx)
					}),
				effects.TargetCreature("target creature token that entered the battlefield this turn",
					effects.IsTokenPredicate(), effects.RfCreatureCEnteredThisTurn())),
		},
	})
}
