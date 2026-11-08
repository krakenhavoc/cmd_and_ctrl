package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tumbleweed Rising — Sorcery {1}{G} (Outlaws of Thunder Junction, #4468):
//
//	"Create an X/X green Elemental creature token, where X is the greatest
//	 power among creatures you control.
//	 Plot {2}{G}"
//
// X is read as the spell resolves (Bouncer's Beatdown's
// b42GreatestPowerControlledBy, layered power), and the token is the green
// Elemental row sized to it. With no creature X is 0, and the 0/0 token is
// created and dies to state-based actions (CR 704.5f), as printed. Plot is
// the ordinary special action (CR 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "ef574346-1d98-447d-acb6-4d8d53c5f651",
		Name:           "Tumbleweed Rising",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{2}{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := b42GreatestPowerControlledBy(ctx.Game, item.Controller)
			t := TokenCard("2/2 green Elemental")
			t.Power, t.Toughness, t.PrintedPTKnown = x, x, true
			return CreateToken{Template: t, N: 1}.Apply(ctx)
		},
	})
}
