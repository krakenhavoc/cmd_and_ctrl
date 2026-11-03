package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sleight of Hand — Sorcery {U} (Secrets of Strixhaven Mystical Archive,
// #4834):
//
//	"Look at the top two cards of your library. Put one of them into your
//	 hand and the other on the bottom of your library."
//
// Impulse's look with two cards: a private look, a mandatory take of one,
// and the other on the bottom (with one left there is no order to choose).
// A library of one card takes it with no prompt, and an empty one does
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05a039ad-1689-4d13-b393-6c1219583b5d",
		Name:         "Sleight of Hand",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return LookAtTopThenTakeOneRestOnBottom(2,
				"Sleight of Hand — put one of them into your hand and the other on the bottom")(ctx.Game, item)
		},
	})
}
