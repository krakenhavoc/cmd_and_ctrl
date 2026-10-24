package effects

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
func vraskasFinalMercyLoseLife(ctx *Context) error {
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
}

func init() {
	Register(Spec{
		OracleID:     "7799413d-e313-4a05-ad28-f4c926c26d28",
		Name:         "Vraska's Final Mercy",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			// The second sentence waits for a prompt the life loss asks (a CR 616 order; CR 608.2c,
			// #4155).
			ModeSteps("You lose 2 life. Destroy target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
				Instruction(vraskasFinalMercyLoseLife),
				Instruction(func(ctx *Context) error {
					t, ok := ModeTarget(ctx, ctx.Occurrence())
					if !ok {
						return nil
					}
					return DestroyTarget{Target: t.ID}.Apply(ctx)
				})),
			ModeSteps("You lose 2 life. Empower Jace 6.", nil,
				Instruction(vraskasFinalMercyLoseLife),
				EmpowerJace{N: 6}),
		),
	})
}
