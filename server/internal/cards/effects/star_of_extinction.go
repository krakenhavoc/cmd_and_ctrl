package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Star of Extinction — Sorcery {5}{R}{R} (Commander Masters, #5948):
//
//	"Destroy target land. Star of Extinction deals 20 damage to each
//	 creature and each planeswalker."
//
// The land is destroyed first, while it is still a legal target (CR
// 608.2b); then the damage is one instance to every creature and every
// planeswalker on the battlefield (damageEachMatching). With its only
// target gone the whole spell does not resolve, so there is no damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "48220e6c-5752-46e0-9b7f-f0eef274d929",
		Name:         "Star of Extinction",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 20},
		},
		Targets: TargetPermanent("target land", Land()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard && ctx.IsTargetLegal(item.Targets[0]) {
				if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return damageEachMatching(ctx, Or(Creature(), Planeswalker()), 20)
		},
	})
}
