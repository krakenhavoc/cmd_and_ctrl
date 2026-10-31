package effects

// Traumatic Critique — Instant {X}{U}{R} (Secrets of Strixhaven, #4456):
//
//	"Traumatic Critique deals X damage to any target. Draw two cards,
//	 then discard a card."
//
// The damage is Stroke of Genius's X read off the announcement. The loot
// is Three Steps Ahead's (drawThenDiscard): the discard is asked once
// both cards are in hand, so a drawn card may be the one discarded, and
// it waits for any question the damage asked. With its only target gone
// the spell does not resolve at all (CR 608.2b), so it draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06c1f8a4-4948-435b-ae98-b1276b57e958",
		Name:         "Traumatic Critique",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetAny(),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: ctx.X()}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 2, 1, "")
			}),
		),
	})
}
