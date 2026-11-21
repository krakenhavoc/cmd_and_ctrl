package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
// vraskasFinalMercyLoseLife is "you lose 2 life".
func vraskasFinalMercyLoseLife(ctx *effects.Context) error {
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
}

func init() {
	effects.Register(effects.Spec{
		OracleID:     "7799413d-e313-4a05-ad28-f4c926c26d28",
		Name:         "Vraska's Final Mercy",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			// The second sentence waits for a prompt the life loss asks (a CR 616 order; CR 608.2c,
			// #4155).
			effects.ModeSteps("You lose 2 life. Destroy target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
				effects.Instruction(vraskasFinalMercyLoseLife),
				effects.Instruction(func(ctx *effects.Context) error {
					t, ok := effects.ModeTarget(ctx, ctx.Occurrence())
					if !ok {
						return nil
					}
					return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
				})),
			effects.ModeSteps("You lose 2 life. Empower Jace 6.", nil,
				effects.Instruction(vraskasFinalMercyLoseLife),
				effects.EmpowerJace{N: 6}),
		),
	})
}
