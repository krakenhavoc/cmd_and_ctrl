package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Overwhelming Stampede — Sorcery {3}{G}{G} (EDHREC top 1,000, #3654):
//
//	"Until end of turn, creatures you control gain trample and get +X/+X,
//	 where X is the greatest power among creatures you control."
//
// Overrun with a computed bonus. X is read once as the spell resolves
// (CR 608.2h), before any creature is pumped, so the bonus is not
// compounded by its own effect, and the set of affected creatures is
// fixed at that moment (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1e96802-fd0c-41f1-aa21-3287d75a0e88",
		Name:         "Overwhelming Stampede",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := b42GreatestPowerControlledBy(ctx.Game, ctx.Controller())
			yours := And(Creature(), YouControl())
			if err := (BoostUntilEOT{
				Match:     yours,
				Power:     x,
				Toughness: x,
				Label:     "Overwhelming Stampede — +X/+X",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    yours,
				Keywords: []string{"trample"},
				Label:    "Overwhelming Stampede — trample",
			}.Apply(ctx)
		},
	})
}
