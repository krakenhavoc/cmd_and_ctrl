package effects

// DealDividedShares deals each still-legal pick of target clause `clause` its
// announced share of a divided amount, as one damage instance, from the
// resolving item's source. A share is not redistributed when a pick has become
// illegal (CR 608.2b). It is DealDividedDamage for a spell whose divided
// clause is not its only one (Fiery Justice).
func DealDividedShares(ctx *Context, clause int) error {
	shares := ctx.Item.Distribution
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, t := range ctx.ClauseTargets(clause) {
			if !ctx.IsTargetLegal(t) || shares[t.ID] <= 0 {
				continue
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: shares[t.ID]}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
