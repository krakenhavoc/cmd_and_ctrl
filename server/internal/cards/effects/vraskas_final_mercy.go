package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vraska's Final Mercy — Sorcery {B}{B} (Reality Fracture):
//
//	"Choose one —
//	 • You lose 2 life. Destroy target creature or planeswalker.
//	 • You lose 2 life. Empower Jace 6."
//
// ADR 0139: the keyword action on a modal bullet. Each bullet loses the
// life first, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7799413d-e313-4a05-ad28-f4c926c26d28",
		Name:         "Vraska's Final Mercy",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("You lose 2 life. Destroy target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2); err != nil {
						return err
					}
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return DestroyTarget{Target: t.ID}.Apply(ctx)
				}),
			ModeDoing("You lose 2 life. Empower Jace 6.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2); err != nil {
						return err
					}
					return EmpowerJace{N: 6}.Apply(ctx)
				}),
		),
	})
}
