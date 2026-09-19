package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reckless Impulse — Sorcery {1}{R} (EDHREC rank 2124):
//
//	"Exile the top two cards of your library. Until the end of your
//	 next turn, you may play those cards."
//
// Two-mana impulse draw with a whole extra turn on the clock, which
// is what makes it a real Divination rather than a gamble: the land
// can be played on either turn and a spell you cannot afford today
// waits for tomorrow. The exile is the S21 impulse grant ("play",
// not "cast", so a land is not stranded); the duration is ADR 0063's
// "until the end of your next turn", keyed on the seat-turn counter
// so it means the same thing from every seat.
//
// Wrenn's Resolve is the same card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "584cf0fd-112a-4ca6-9c0e-1f3228a7d325",
		Name:         "Reckless Impulse",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b19ExileTopTwoUntilEndOfNextTurn(ctx.Game, item)
		},
	})
}
