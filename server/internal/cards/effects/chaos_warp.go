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
// #783: the shuffle and the reveal are the tuck's CONTINUATION, not the
// next two lines. CR 608.2c runs the instructions in order and the tuck
// can PAUSE — a commander's owner is asked about the command zone —
// so writing them on the next line shuffled and revealed while the
// permanent was still on the battlefield, and an owner who then
// declined had their commander land on TOP of the already-shuffled
// library instead of shuffled in. Now nothing happens until the
// question is answered.
//
// The answer is deliberately IGNORED. "Shuffles it into their library,
// THEN reveals the top card" is one sentence about the library, not an
// "if you do": a commander that goes to the command zone instead still
// leaves its owner shuffling and revealing, because the shuffle is not
// conditional on where the permanent ended up.
func init() {
	Register(Spec{
		OracleID:     "07a0cba9-8768-4fd9-a3d5-b0f83b4bf8e8",
		Name:         "Chaos Warp",
		Completeness: CompletenessFull,
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
			source := ctx.Source()
			// The continuation takes the live *Game rather than
			// capturing one, the contract every continuation in the
			// engine follows: an undo restores this game's fields in
			// place, so a captured *Game would be the wrong one.
			return ctx.Game.TuckToLibraryThenForEffect(id, game.TuckOptions{}, func(g *game.Game, _ bool) error {
				if err := g.ShuffleLibraryForEffect(owner); err != nil {
					return err
				}
				_, err := revealTopThenPutIfMatch(g, source, owner, chaosWarpAnyPermanent,
					false, "Chaos Warp — revealed from the top of the library")
				return err
			})
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
