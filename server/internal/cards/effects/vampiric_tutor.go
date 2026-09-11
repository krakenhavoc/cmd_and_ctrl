package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampiric Tutor — Instant {B}:
//
//	"Search your library for a card, then shuffle and put that card
//	 on top of your library. You lose 2 life."
//
// S22: the put-on-top is real. This shipped in S14 sending the
// tutored card straight to HAND with a note saying the difference
// "matters when racing or dodging discard" and would be revisited.
// It is revisited: SearchLibrary.ToTop places the card after the
// shuffle, which is the whole clause — the card is a known quantity
// sitting on an unknown library, and it costs you a draw step to
// collect it.
//
// That card-disadvantage is what makes this a one-mana tutor instead
// of Demonic Tutor's three, so modelling it as a to-hand search was
// quietly printing a much better card.
//
// No reveal: Vampiric Tutor does not say "reveal it", so only the
// controller knows what is on top.
func init() {
	Register(Spec{
		OracleID: "ededbdae-d9dc-4206-9335-d7158f2d7700",
		Name:     "Vampiric Tutor",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (SearchLibrary{
				Player:  ctx.Controller(),
				Dest:    game.ZoneLibrary,
				ToTop:   true,
				Limit:   1,
				Reveal:  false,
				Shuffle: true,
				Reason:  "Vampiric Tutor — search your library for a card",
			}).Apply(ctx); err != nil {
				return err
			}
			// The life loss is a separate sentence and does not wait
			// on the search, so it stays inline rather than riding
			// the Then continuation.
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
		},
	})
}
