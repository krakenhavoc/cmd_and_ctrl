package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	end := AtYourEndStep("Abiding Grace — choose one", func(*game.Game, *game.StackItem) error { return nil })
	end.Modes = ChooseOne(
		ModeWithPurpose(ModeSteps("You gain 1 life.", nil, GainLife{Amount: 1}), game.Purpose{LifeGain: 1}),
		ModeDoing("Return target creature card with mana value 1 from your graveyard to the battlefield.",
			TargetCardInGraveyard("target creature card with mana value 1 from your graveyard",
				Creature(), YouOwn(), ManaValueGE(1), ManaValueLE(1)),
			func(_ *game.StackItem, ctx *Context, _ int) error {
				t, ok := ModeTarget(ctx, 0)
				if !ok {
					return nil
				}
				return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "f4e7bc58-336d-4670-84cb-a46ca1f48671",
		Name:         "Abiding Grace",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{end},
	})
}
