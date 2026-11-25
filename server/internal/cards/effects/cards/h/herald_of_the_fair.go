package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of the Fair — Creature — Human {2}{W}, 3/2 (Kaladesh, #5539):
//
//	"When this creature enters, target creature you control gets +1/+1 until
//	 end of turn."
//
// Yotian Frontliner's targeted pump, without "another": the Herald may
// target itself, and the target is chosen as the trigger goes on the stack.
// The pump is skipped if the target has left or is no longer legal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6bb7d985-b033-46a2-9bbb-6f7ee62aaaff",
		Name:         "Herald of the Fair",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Herald of the Fair — target creature you control gets +1/+1 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Herald of the Fair — +1/+1"}.Apply(ctx)
					}),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1))),
		},
	})
}
