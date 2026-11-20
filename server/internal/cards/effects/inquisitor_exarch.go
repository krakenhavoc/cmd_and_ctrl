package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inquisitor Exarch — Creature — Phyrexian Cleric {W}{W}, 2/2 (New Phyrexia, #5575):
//
//	"When this creature enters, choose one —
//	 • You gain 2 life.
//	 • Target opponent loses 2 life."
//
// A modal enters trigger (Apothecary Stomper's shape): the bullet is
// chosen as the trigger goes on the stack (CR 603.3c), and the opponent
// bullet targets (CR 603.3d).
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Inquisitor Exarch — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("You gain 2 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
		}),
		ModeDoing("Target opponent loses 2 life.", TargetPlayer("target opponent", Opponent()),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), t.ID, -2)
			}),
	)
	Register(Spec{
		OracleID:     "7deae378-9291-487e-9f0b-1deaf37218a8",
		Name:         "Inquisitor Exarch",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
