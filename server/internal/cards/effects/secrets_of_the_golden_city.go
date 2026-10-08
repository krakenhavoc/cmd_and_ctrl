package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secrets of the Golden City — Sorcery {1}{U}{U} (EDHREC rank 15017):
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 Draw two cards. If you have the city's blessing, draw three cards
//	 instead."
//
// Ascend on a sorcery is a spell ability (CR 702.131a): the engine
// checks as the spell resolves, BEFORE its other instructions
// (ascendSpellLocked in game), so the clause below reads the answer
// the spell has just earned. A player who already had the blessing
// reads it too, from the designation, however few permanents they
// control now.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c0dda0d0-1fae-4777-ba86-9fe7990bf3a8",
		Name:            "Secrets of the Golden City",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 2
			if YouHaveTheCitysBlessing(ctx.Game, item.Controller) {
				n = 3
			}
			return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
		},
	})
}
