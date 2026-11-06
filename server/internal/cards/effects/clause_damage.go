package effects

// clause_damage.go — "deals A damage to <first target> and B damage to
// <second target>" as one damage instance (CR 120.2): Arc Trail, Conduct
// Electricity.

// damageEachClauseTarget deals amounts[slot] damage from the resolving
// spell to the pick that answered each target clause, all at once. A
// clause with no pick, or whose pick has become illegal (CR 608.2b), is
// skipped.
func damageEachClauseTarget(ctx *Context, amounts ...int) error {
	return ctx.Game.DamageInstanceForEffect(func() error {
		for slot, n := range amounts {
			t, ok := ctx.ClauseTarget(slot)
			if !ok {
				continue
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
