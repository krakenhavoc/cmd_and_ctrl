package effects

// Kick in the Door — Sorcery {R}:
//
//	"Put a +1/+1 counter on target creature. That creature gains haste
//	 until end of turn and can't be blocked by Walls this turn. Venture
//	 into the dungeon."
//
// "Can't be blocked by Walls this turn" is a cantBeBlockedBySubtypes
// block rule pinned to the creature (#4100), read against the blocker's
// live creature types, so a changeling is a Wall.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d3278ddb-c198-450d-946e-8547e5c96dd1",
		Name:         "Kick in the Door",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		// Each sentence waits for the prompt the one before asks (a CR 616 order
		// on the counter; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return AddCounter{Target: id, Kind: "+1/+1", N: 1}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Kick in the Door — haste until end of turn"}.Apply(ctx)
				}
				return nil
			}),
			Instruction(func(ctx *Context) error {
				if id, ok := b16FirstLegalTargetCard(ctx); ok {
					return CantBeBlockedThisTurnBySubtypes{Target: id, Subtypes: []string{ctx.TextSubtype("Wall")}, Text: "Walls", Label: "Kick in the Door — can't be blocked by Walls this turn"}.Apply(ctx)
				}
				return nil
			}),
			VentureIntoTheDungeon{},
		),
	})
}
