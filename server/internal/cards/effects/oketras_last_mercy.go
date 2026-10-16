package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oketra's Last Mercy — Sorcery {1}{W}{W}:
//
//	"Your life total becomes equal to your starting life total. Lands
//	 you control don't untap during your next untap step."
//
// The life total becomes the table's starting total (40 in Commander)
// by gaining or losing the difference (CR 119.5), so "can't gain life"
// stops the gain and a life-gain payoff sees it. Then #3995's
// player-wide skip over the caster's lands.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5e0aa12-920d-4700-a287-5ca24933abb1",
		Name:         "Oketra's Last Mercy",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := b31LifeBecomes(ctx, ctx.Controller(), startingLifeOf(ctx.Game)); err != nil {
				return err
			}
			return LandsYouControlDontUntapNextStep(ctx)
		},
	})
}
