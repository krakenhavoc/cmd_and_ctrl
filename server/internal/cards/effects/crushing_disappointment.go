package effects

// Crushing Disappointment — Instant {3}{B}:
//
//	"Each player loses 2 life. You draw two cards."
//
// Risky Shortcut's two sentences in the other order, at instant speed:
// every live player, the caster included, loses 2 life first, and only
// then does the caster draw. The order is the printed one and it
// matters: a caster at 2 life is dead at the next state check whatever
// the draws would have found. Life LOSS, not damage, so no prevention
// shield or lifelink sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77a1958e-ee24-4305-9dad-da8372803be8",
		Name:         "Crushing Disappointment",
		Completeness: CompletenessFull,
		// The draw waits for the prompts the life losses ask (a CR 616 order; CR
		// 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, id := range tablePlayers(ctx) {
					if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), id, -2); err != nil {
						return err
					}
				}
				return nil
			}),
			DrawCards{N: 2},
		),
	})
}
