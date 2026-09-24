package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impulse — Instant {1}{U}:
//
//	"Look at the top four cards of your library. Put one of them into
//	 your hand and the rest on the bottom of your library in any order."
//
// The instant-speed dig the whole "look at N, take one, the rest on
// the bottom in any order" family is built from (Anticipate, Stock Up,
// Dig Through Time, Teferi, Who Slows the Sunset's −2). It is three
// engine pieces in a row: a private LOOK (only the caster becomes a
// knower), a TakeFromLibraryToHand with Max 1 and no "may" — "put one
// of them" is mandatory — and TakeRestOnBottomInAnyOrder, the ordered
// bottom (#996, ADR 0088). The card taken is not revealed; Impulse
// does not say so.
//
// A library of one card is taken with no prompt (the only answer), and
// an empty library does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6bd2902-7f8b-419e-bbc7-bcab0c1b7e01",
		Name:         "Impulse",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return LookAtTopThenTakeOneRestOnBottom(4, "Impulse — put one of them into your hand")(ctx.Game, item)
		},
	})
}
