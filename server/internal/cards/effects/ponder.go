package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ponder — Sorcery {U}:
//
//	"Look at the top three cards of your library, then put them back
//	 in any order. You may shuffle."
//	"Draw a card."
//
// Preordain's twin, one card deeper and one card less selective: you
// see three but cannot bin any of them, so the arrangement is the
// whole decision — unless all three are bad, which is what the shuffle
// is for.
//
// # Two prompts, in order
//
// The card is a chain, and it has to be one. "Put them back in any
// order" is answered first, "you may shuffle" second, and the second
// question only makes sense once the player has seen the answer to the
// first — a player who has just arranged three keepers says no, and a
// player looking at three lands says yes and throws the arrangement
// away. Asking both at once would be asking the player to shuffle
// before they had looked.
//
// So the shuffle prompt is queued by the look-at-top prompt's
// continuation, which is the composition the choice queue gained for
// exactly this (see game/chained_choice.go). LookAtTop's "then" no longer
// draws; it asks. It is a registered tail (#4301), so a table waiting on the
// look-at is a restore point as the shuffle question after it is. The draw moves down one link, onto both branches of
// the confirm, because "then draw a card" is what happens after the
// whole reorder-and-maybe-shuffle clause finishes — a draw on only one
// branch would be a card that Ponder sometimes forgets to draw.
//
// The ordering is the entire card: the draw must come after the
// shuffle, or it is a card off the arrangement the player just
// discarded, and it must come after the reorder, or it is a card the
// player is still deciding about.
func init() {
	Register(Spec{
		OracleID:     "02090581-61aa-4348-ad57-451be8ee91c2",
		Name:         "Ponder",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return LookAtTop{
				Player:   controller,
				N:        3,
				ThenTail: ctx.ItemTail(ponderAskShuffle),
			}.Apply(ctx)
		},
	})
}

// ponderAskShuffle is the look-at's "then": ask whether to shuffle (#4301:
// a registered tail, with the asking item as its data). An on-disk key.
var ponderAskShuffle = CardTail("ponder/ask-shuffle", func(ctx *Context, _ game.TailOutcome) error {
	return MayChoice{
		Question: "Ponder — shuffle your library?",
		YesLabel: "Shuffle",
		NoLabel:  "Keep that order",
		Then:     ponderShuffle,
	}.Apply(ctx)
})

// ponderShuffle is "you may shuffle", then the draw: the draw is on both
// branches (#4298: keyed, so a table waiting on the question is a restore
// point).
var ponderShuffle = MayAnswer("confirm/ponder-shuffle", func(ctx *Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := ctx.Game.ShuffleLibraryForEffect(r.Chooser); err != nil {
			return err
		}
	}
	return ctx.Game.DrawNForEffect(r.Chooser, 1)
})
