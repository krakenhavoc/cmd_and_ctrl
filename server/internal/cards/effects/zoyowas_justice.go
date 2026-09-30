package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zoyowa's Justice — Instant {1}{R}:
//
//	"The owner of target artifact or creature with mana value 1 or
//	 greater shuffles it into their library. Then that player discovers
//	 X, where X is its mana value."
//
// The discoverer is the target's OWNER, not the caster: the prompt, the
// free cast and the hand all belong to them. X is the permanent's mana
// value as it sat on the battlefield, read before it leaves (CR 608.2h).
// The tuck can pause for a commander's CR 903.9 offer, so the shuffle
// and the discover wait in its continuation; the discover happens
// whether or not the permanent reached the library ("then", not "if
// you do"). Discover is ADR 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:     "17e11f0c-4ac7-4060-9e0b-27cd5c151e50",
		Name:         "Zoyowa's Justice",
		Completeness: CompletenessFull,
		Discovers:    true,
		Targets: TargetPermanent("target artifact or creature with mana value 1 or greater",
			Or(Artifact(), Creature()), ManaValueGE(1)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			owner := c.Owner
			x, _ := ctx.Game.ManaValueForEffect(c)
			return ctx.Game.TuckToLibraryThenForEffect(id, game.TuckOptions{}, func(g *game.Game, _ bool) error {
				if err := g.ShuffleLibraryForEffect(owner); err != nil {
					return err
				}
				return Discover{Player: owner, N: x}.Apply(NewContext(g, item))
			})
		},
	})
}
