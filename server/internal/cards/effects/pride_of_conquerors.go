package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pride of Conquerors — Instant {1}{W}:
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 Creatures you control get +1/+1 until end of turn. If you have the
//	 city's blessing, those creatures get +2/+2 until end of turn
//	 instead."
//
// The blessing is read as the spell resolves, and the creatures are
// snapshotted then (CR 611.2c): one that arrives later gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e45a1c99-a020-49ab-8971-35c6bcd096c8",
		Name:            "Pride of Conquerors",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 1
			if YouHaveTheCitysBlessing(ctx.Game, item.Controller) {
				n = 2
			}
			return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: n, Toughness: n, Label: "Pride of Conquerors"}.Apply(ctx)
		},
	})
}
