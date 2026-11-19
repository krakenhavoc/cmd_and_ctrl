package effects

import "github.com/google/uuid"

// s82BnsB09DamageEachOpponentBy deals each opponent of the item's controller
// damage equal to n(opp), all as one damage instance (CR 120.3). An opponent
// whose amount is zero or less is dealt nothing.
func s82BnsB09DamageEachOpponentBy(ctx *Context, n func(opp uuid.UUID) int) error {
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, opp := range ctx.Opponents() {
			if amt := n(opp); amt > 0 {
				if err := ctx.Game.DealDamageToPlayerForEffect(ctx.Source(), opp, amt); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
