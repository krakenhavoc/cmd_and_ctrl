package effects

// Prophetic Bolt — Instant {3}{U}{R} (Apocalypse, #5614):
//
//	"Prophetic Bolt deals 4 damage to any target. Look at the top four
//	 cards of your library. Put one of those cards into your hand and
//	 the rest on the bottom of your library in any order."
//
// Lightning Bolt's body at 4 (damageToFirstTarget), then Impulse
// (LookAtTopThenTakeOneRestOnBottom). The two sentences are two steps
// (ResolveSteps, CR 608.2c), so the look waits for anything the damage
// asks first. A Bolt whose only target is gone does not resolve at all
// (CR 608.2b), so it looks at nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d5e4bf5e-1a66-4315-8c82-640c27977b88",
		Name:         "Prophetic Bolt",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 4)),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return damageToFirstTarget(4)(ctx.Item, ctx)
			}),
			Instruction(func(ctx *Context) error {
				return LookAtTopThenTakeOneRestOnBottom(4, "Prophetic Bolt — put one of them into your hand")(ctx.Game, ctx.Item)
			}),
		),
	})
}
