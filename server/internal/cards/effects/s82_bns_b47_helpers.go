package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// bnsB47DamageToAnyTarget is the Effect body of "<this> deals N damage to any
// target" for a spell with one any-target clause (Open Fire). The source is
// the spell itself; a target that has gone leaves nothing to damage
// (CR 608.2b).
func BnsB47DamageToAnyTarget(amount int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if len(item.Targets) == 0 {
			return nil
		}
		ctx := NewContext(g, item)
		return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: amount}.Apply(ctx)
	}
}
