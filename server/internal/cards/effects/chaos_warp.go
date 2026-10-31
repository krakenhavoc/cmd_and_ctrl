package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: shuffles a permanent away
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
			// The rest of the card is the tuck's keyed tail (#4253 slice
			// 10a): the owner rides it as data.
			return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{},
				withCarry(ctx.Tail(chaosWarpShuffleReveal), c.Owner))
		},
	})
}

// chaosWarpShuffleReveal is "the owner shuffles, then reveals the top card
// of their library; if it's a permanent card, they put it onto the
// battlefield": Carry[0] is the owner, Source the Chaos Warp.
var chaosWarpShuffleReveal = game.RegisterTail("then/chaos-warp-shuffle-reveal", func(g *game.Game, o game.TailOutcome) error {
	owner := tailCarry0(o)
	if owner == uuid.Nil {
		return nil
	}
	if err := g.ShuffleLibraryForEffect(owner); err != nil {
		return err
	}
	// #1997: "they put it onto the battlefield" — the owner puts it
	// (CR 110.2a), which Arboria reads.
	return g.PutByEnteringPlayerForEffect(func() error {
		_, err := revealTopThenPutIfMatch(g, o.Tail.Source, owner, chaosWarpAnyPermanent,
			false, "Chaos Warp — revealed from the top of the library")
		return err
	})
})

// chaosWarpAnyPermanent is "if it's a permanent card"; the reveal
// helper already refuses nonpermanents and tokens, so every card it
// asks about qualifies. The token half matters here more than anywhere:
// Chaos Warp on a token tucks it, and without the refusal a token that
// came back up as the top card would be put straight back onto the
// battlefield (CR 111.8 says it can't be).
func chaosWarpAnyPermanent(game.Card) bool { return true }
