package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bilbo, Fellow Conspirator — Legendary Creature — Halfling Citizen {2}{G},
// 2/3 (The Hobbit Eternal, HOC):
//
//	"If you would create a Food token, instead create a Food token and a
//	 Treasure token."
//
// Academy Manufactor's replacement with a smaller list: the CR 701.7b
// creation event carries groups, and the Food group becomes a Food group
// and a Treasure group of the same count, so "create two Foods" is two of
// each. Groups of other kinds in the same instruction are left alone. It
// watches only tokens created under Bilbo's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d273f87d-1271-49d2-a628-2d1bd888f56c",
		Name:         "Bilbo, Fellow Conspirator",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			hobA9InsteadCreateTokens("Bilbo, Fellow Conspirator: a Food and a Treasure instead",
				func(t game.Card) bool { return t.HasSubtype("Food") },
				func() []game.Card { return []game.Card{FoodToken(), TreasureToken()} }),
		},
	})
}
