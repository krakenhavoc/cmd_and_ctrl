package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abiding Grace — Enchantment {2}{W} (Modern Horizons 3):
//
//	"At the beginning of your end step, choose one —
//	 • You gain 1 life.
//	 • Return target creature card with mana value 1 from your graveyard to
//	   the battlefield."
//
// A modal end-step trigger (Adaptive Sporesinger's shape, CR 603.3c): the mode
// is chosen as the trigger goes on the stack and only the second bullet
// targets, a creature card with mana value exactly 1 in your own graveyard,
// read again as it resolves (CR 608.2b). It returns under your control.
//
// No simplification.
func init() {
	end := effects.AtYourEndStep("Abiding Grace — choose one", func(*game.Game, *game.StackItem) error { return nil })
	end.Modes = effects.ChooseOne(
		effects.ModeWithPurpose(effects.ModeSteps("You gain 1 life.", nil, effects.GainLife{Amount: 1}), game.Purpose{LifeGain: 1}),
		effects.ModeDoing("Return target creature card with mana value 1 from your graveyard to the battlefield.",
			effects.TargetCardInGraveyard("target creature card with mana value 1 from your graveyard",
				effects.Creature(), effects.YouOwn(), effects.ManaValueGE(1), effects.ManaValueLE(1)),
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				t, ok := effects.ModeTarget(ctx, 0)
				if !ok {
					return nil
				}
				return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "f4e7bc58-336d-4670-84cb-a46ca1f48671",
		Name:         "Abiding Grace",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{end},
	})
}
