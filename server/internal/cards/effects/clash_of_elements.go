package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clash of Elements — Instant {1}{U}{R}:
//
//	"Choose target nonland permanent. Its owner may put it on top of
//	 their library. If they do, Clash of Elements deals 2 damage to
//	 them. If they didn't put the card on top of their library, they
//	 put it on the bottom."
//
// The question goes to the permanent's OWNER, not the caster. "Yes"
// tucks it on top and, only if it actually reached a library, deals 2
// damage to that owner (a commander whose owner takes the command-zone
// offer instead, CR 903.9, is not damaged). "No" tucks it on the
// bottom. A token ceases to exist either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78304a06-4c9a-4cdf-b67f-067d88762381",
		Name:         "Clash of Elements",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 || ts[0].Kind != game.TargetCard {
				return nil
			}
			id := ts[0].ID
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			owner := c.Owner
			return MayChoice{
				Player:   owner,
				Question: "Clash of Elements — put " + c.Name + " on top of your library? (Yes: take 2 damage. No: it goes on the bottom.)",
				YesLabel: "Top (take 2 damage)",
				NoLabel:  "Bottom",
				OnYes: func(ctx *Context) error {
					return ctx.Game.TuckToLibraryThenForEffect(id, game.TuckOptions{}, func(g *game.Game, tucked bool) error {
						if !tucked {
							return nil
						}
						return DealDamage{Source: item.SourceCardID, Target: owner, Amount: 2}.Apply(NewContext(g, item))
					})
				},
				OnNo: func(ctx *Context) error {
					return ctx.Game.TuckToLibraryForEffect(id, true)
				},
			}.Apply(ctx)
		},
	})
}
