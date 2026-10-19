package effects

// Riot Control — Instant {2}{W}:
//
//	"You gain 1 life for each creature your opponents control. Prevent
//	 all damage that would be dealt to you this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the life is counted as the spell
// resolves, then the not-one-use shield protects you, from every source,
// for the rest of the turn.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "a189e20f-1072-4608-9e50-da2a9d7a7ee3",
		Name:         "Riot Control",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			// The shield waits for a prompt the life gain asks (CR 608.2c, #4155).
			Instruction(func(ctx *Context) error {
				ctx.Game.RecomputeLayersIfStaleLocked()
				n := 0
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.IsCreature() && c.Controller != ctx.Controller() {
						n++
					}
				}
				return GainLife{Amount: n}.Apply(ctx)
			}),
			PreventDamageFromSource{Protect: ShieldYou},
		),
	})
}
