package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chaos Warp — Instant {2}{R}:
//
//	"The owner of target permanent shuffles it into their library,
//	 then reveals the top card of their library. If it's a permanent
//	 card, they put it onto the battlefield."
//
// Red's catch-all removal, paid for with a lottery ticket the victim
// holds. Every step is the OWNER's: the permanent goes into its
// owner's library, that library is shuffled and revealed from, and the
// card that comes up enters under its owner's control — a stolen
// creature warped away replaces itself for the player it was stolen
// from, not for the thief.
//
// The shuffle-in is the ordinary tuck-then-shuffle (a commander's
// owner is offered the command zone, CR 903.9); the reveal and the put
// are #745's library-to-battlefield move, so a land that comes up is
// not a land drop (CR 305.4) and the entry runs through the CR 614
// pipeline.
//
// DECLARED SIMPLIFICATION, weaker for the caster: a commander's owner
// is asked about the command zone (CR 903.9), and that prompt pauses
// the tuck while the shuffle and the reveal carry on. An owner who
// then keeps the commander in the library has it land on TOP of the
// already-shuffled library rather than shuffled in: the shuffle cannot
// wait for an answer without a continuation frame the tuck does not
// have. The victim knows their next draw, which only ever helps them.
func init() {
	Register(Spec{
		OracleID:     "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8",
		Name:         "Chaos Warp",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"If a commander is warped and its owner chooses not to put it into the command zone, it goes on top of their library instead of being shuffled in."},
		Targets:      TargetPermanent("target permanent", Permanent()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			owner := c.Owner
			if err := ctx.Game.TuckToLibraryForEffect(id, false); err != nil {
				return err
			}
			if err := ctx.Game.ShuffleLibraryForEffect(owner); err != nil {
				return err
			}
			_, err := revealTopThenPutIfMatch(ctx.Game, ctx.Source(), owner, chaosWarpAnyPermanent,
				false, "Chaos Warp — revealed from the top of the library")
			return err
		},
	})
}

// chaosWarpAnyPermanent is "if it's a permanent card"; the reveal
// helper already refuses nonpermanents and tokens, so every card it
// asks about qualifies. The token half matters here more than anywhere:
// Chaos Warp on a token tucks it, and without the refusal a token that
// came back up as the top card would be put straight back onto the
// battlefield (CR 111.8 says it can't be).
func chaosWarpAnyPermanent(game.Card) bool { return true }
